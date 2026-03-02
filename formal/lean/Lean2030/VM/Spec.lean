import Std

namespace Lean2030.VM

abbrev HashFn := List UInt8 → List UInt8 → List UInt8

def defaultHash : HashFn := fun a b => a ++ b

def defaultMaxCycles : Nat := 1 <<< 20

structure Config where
  maxCycles : Nat := defaultMaxCycles

namespace ExecError
  inductive T where
    | emptyProgram
    | cycleLimit
    | stackUnderflow
    | invalidOperand
  deriving BEq, Repr
end ExecError

abbrev Value := List UInt8

inductive Opcode where
  | add
  | mul
  | hash
  | verify
  | push (idx : Nat)
  | dup
  deriving Repr, BEq, DecidableEq

abbrev Program := List Opcode

instance : Inhabited Opcode := ⟨Opcode.add⟩

structure RunState where
  stack : List Value := []
  gasUsed : Nat := 0
  steps : Nat := 0
  deriving Repr, BEq

instance : Inhabited RunState := ⟨{}⟩

abbrev InputTape := List Value

def listGet? (stack : List α) (i : Nat) : Option α :=
  match stack with
  | [] => none
  | x :: xs =>
      match i with
      | 0 => some x
      | i + 1 => listGet? xs i

open ExecError

-- Gas schedule copied from the Go reference implementation.
def gasAdd : Nat := 3

def gasMul : Nat := 5

def gasHash : Nat := 30

def gasVerify : Nat := 50

def gasPush : Nat := 2

def gasDup : Nat := 2

def gasOf (op : Opcode) : Nat :=
  match op with
  | Opcode.add => gasAdd
  | Opcode.mul => gasMul
  | Opcode.hash => gasHash
  | Opcode.verify => gasVerify
  | Opcode.push _ => gasPush
  | Opcode.dup => gasDup

/-- Additive operator used by LeanVM ADD.
    This is a byte-wise XOR, matching the current Go toy definition.
-/
def addBytes : Value → Value → Value
  | [], [] => []
  | x :: xs, [] => x :: addBytes xs []
  | [], y :: ys => y :: addBytes [] ys
  | x :: xs, y :: ys => (x ^^^ y) :: addBytes xs ys

/-- Multiplicative operator used by LeanVM MUL.
    This is a hash combinator placeholder in the reference implementation.
-/
def mulBytes (hash : HashFn) (a b : Value) : Value :=
  hash a b

open Nat

private def pop1 : RunState → Option (Value × RunState)
  | ⟨[], gasUsed, steps⟩ => none
  | ⟨top :: rest, gasUsed, steps⟩ => some (top, ⟨rest, gasUsed, steps⟩)

private def pop2 : RunState → Option (Value × Value × RunState)
  | ⟨[], gasUsed, steps⟩ => none
  | ⟨[_], gasUsed, steps⟩ => none
  | ⟨top1 :: top2 :: rest, gasUsed, steps⟩ => some (top2, top1, ⟨rest, gasUsed, steps⟩)

/-- Execute one opcode under a pre-charged step budget.
- We assume `steps` has already been incremented in the caller.
-/
def execOp (hash : HashFn) (inputs : InputTape) (s : RunState) (op : Opcode) : Except ExecError.T RunState :=
  match op with
  | Opcode.add =>
      match pop2 s with
      | none => Except.error ExecError.T.stackUnderflow
      | some (a, b, next) =>
          Except.ok { next with
            stack := addBytes a b :: next.stack
            gasUsed := next.gasUsed + gasAdd }

  | Opcode.mul =>
      match pop2 s with
      | none => Except.error ExecError.T.stackUnderflow
      | some (a, b, next) =>
          Except.ok { next with
            stack := mulBytes hash a b :: next.stack
            gasUsed := next.gasUsed + gasMul }

  | Opcode.hash =>
      match pop1 s with
      | none => Except.error ExecError.T.stackUnderflow
      | some (v, next) =>
          Except.ok { next with
            stack := mulBytes hash v v :: next.stack
            gasUsed := next.gasUsed + gasHash }

  | Opcode.verify =>
      match pop2 s with
      | none => Except.error ExecError.T.stackUnderflow
      | some (a, b, next) =>
          let flag : UInt8 := if a == b then 1 else 0
          Except.ok { next with
            stack := [flag] :: next.stack
            gasUsed := next.gasUsed + gasVerify }

  | Opcode.push idx =>
      match listGet? inputs idx with
      | none => Except.error ExecError.T.invalidOperand
      | some value =>
          Except.ok { s with
            stack := value :: s.stack
            gasUsed := s.gasUsed + gasPush }

  | Opcode.dup =>
      match s.stack with
      | [] => Except.error ExecError.T.stackUnderflow
      | top :: rest =>
          Except.ok { s with
            stack := top :: top :: rest
            gasUsed := s.gasUsed + gasDup }

/-- Normalize Go-compatible defaults: if maxCycles is set to 0, it means default.
-/
def normalize (cfg : Config) : Nat :=
  if cfg.maxCycles = 0 then defaultMaxCycles else cfg.maxCycles

/-- Execute a LeanVM program.
    On success the output is the top stack element if any exists.
-/
def run (hash : HashFn) (inputs : InputTape) (cfg : Config) (prog : Program) :
    Except ExecError.T (Option Value × RunState) :=
  if prog = [] then
    Except.error ExecError.T.emptyProgram
  else
    let maxCycles := normalize cfg
    let rec loop (ops : Program) (s : RunState) : Except ExecError.T RunState :=
      match ops with
      | [] => Except.ok s
      | op :: rest =>
          if s.steps >= maxCycles then
            Except.error ExecError.T.cycleLimit
          else
            let s' : RunState := { s with steps := s.steps + 1 }
            match execOp hash inputs s' op with
            | Except.error e => Except.error e
            | Except.ok okState => loop rest okState
    match loop prog {} with
    | Except.error e => Except.error e
    | Except.ok finalState =>
        Except.ok (listGet? finalState.stack 0, finalState)

end Lean2030.VM
