import Std

namespace Lean2030.EVM

abbrev Word := Nat

def two256 : Nat := 1 <<< 256

def toWord : Nat → Word := fun n => n % two256

def two255 : Nat := 1 <<< 255

def opAdd : Word → Word → Word := fun x y => toWord (x + y)

def opSub : Word → Word → Word := fun x y => toWord (x + (two256 - (y % two256)))

def opMul : Word → Word → Word := fun x y => toWord (x * y)

def opDiv : Word → Word → Word := fun x y => if y = 0 then 0 else x / y

def opMod : Word → Word → Word := fun x y => if y = 0 then 0 else x % y

def opEq : Word → Word → Word := fun x y => if x = y then 1 else 0

def opIsZero : Word → Word := fun x => if x = 0 then 1 else 0

def opAnd : Word → Word → Word := fun x y => x &&& y

def opOr : Word → Word → Word := fun x y => x ||| y

def opXor : Word → Word → Word := fun x y => x ^^^ y

def opNot : Word → Word := fun x => toWord (two256 - 1 - x)

def opLt : Word → Word → Word := fun x y => if x < y then 1 else 0

def opGt : Word → Word → Word := fun x y => if x > y then 1 else 0

abbrev Stack := List Word

structure Config where
  maxCycles : Nat := 1 <<< 20

inductive Error where
  | emptyProgram
  | cycleLimit
  | stackUnderflow
  | invalidOperand
  | unsupported
  deriving BEq, Repr, DecidableEq

inductive Op where
  | stop
  | add
  | mul
  | sub
  | div
  | mod
  | eq
  | isZero
  | andOp
  | orOp
  | xorOp
  | notOp
  | lt
  | gt
  | push (value : Word)
  | dup (n : Nat)
  | swap (n : Nat)
  | pop
  | jumpdest
  | jump
  | jumpi
  deriving Repr, BEq, DecidableEq

abbrev Program := List Op

structure State where
  stack : Stack := []
  gasUsed : Nat := 0
  steps : Nat := 0
  deriving Repr, BEq, DecidableEq

instance : Inhabited State := ⟨{}⟩

def gasOf : Op → Nat
  | Op.stop => 0
  | Op.add | Op.mul | Op.sub | Op.div | Op.mod => 3
  | Op.eq | Op.isZero | Op.andOp | Op.orOp | Op.xorOp | Op.notOp | Op.lt | Op.gt => 3
  | Op.push _ => 3
  | Op.dup _ => 3
  | Op.swap _ => 3
  | Op.pop | Op.jumpdest | Op.jump | Op.jumpi => 3

def normalizeCycles (cfg : Config) : Nat :=
  if cfg.maxCycles = 0 then (1 <<< 20) else cfg.maxCycles

/-- Pop `n` elements from the top of stack (front of list), returning popped values and remainder. -/
def popn : Nat → List α → Option (List α × List α)
  | 0, xs => some ([], xs)
  | (_ + 1), [] => none
  | n + 1, x :: xs =>
      match popn n xs with
      | none => none
      | some (taken, rest) => some (x :: taken, rest)

def pop1 (stack : List α) : Option (α × List α) :=
  match stack with
  | [] => none
  | x :: xs => some (x, xs)

def pop2 (stack : List α) : Option (α × α × List α) :=
  match stack with
  | x :: y :: xs => some (x, y, xs)
  | _ => none

def listGet? (stack : List α) (i : Nat) : Option α :=
  match stack with
  | [] => none
  | x :: xs =>
      match i with
      | 0 => some x
      | i + 1 => listGet? xs i

def nthFromTop (stack : List Word) (n : Nat) : Option Word :=
  listGet? stack n

def setNthFromTop (stack : List Word) (n : Nat) (v : Word) : Option (List Word) :=
  match n, stack with
  | 0, [] => none
  | 0, x :: xs => some (v :: xs)
  | n + 1, [] => none
  | n + 1, x :: xs =>
      match setNthFromTop xs n v with
      | none => none
      | some xs' => some (x :: xs')

def swapTop (stack : List Word) (n : Nat) : Option (List Word) :=
  if h : n = 0 then
    none
  else
    match listGet? stack 0, listGet? stack n with
    | none, _ => none
    | _, none => none
    | some a, some b =>
        match setNthFromTop stack 0 b with
        | none => none
        | some atTop =>
            match setNthFromTop atTop n a with
            | none => none
            | some swapped => some swapped

def withGas (s : State) (op : Op) : State :=
  { s with gasUsed := s.gasUsed + gasOf op, steps := s.steps + 1 }

def exec (s : State) : Op → Except Error State
  | Op.stop => Except.ok s
  | Op.push v => Except.ok (withGas { s with stack := v :: s.stack } (Op.push v))
  | Op.add =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opAdd x y :: rest } Op.add)
  | Op.sub =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opSub x y :: rest } Op.sub)
  | Op.mul =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opMul x y :: rest } Op.mul)
  | Op.div =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opDiv x y :: rest } Op.div)
  | Op.mod =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opMod x y :: rest } Op.mod)
  | Op.eq =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opEq x y :: rest } Op.eq)
  | Op.isZero =>
      match pop1 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, rest) =>
          Except.ok (withGas { s with stack := opIsZero x :: rest } Op.isZero)
  | Op.andOp =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opAnd x y :: rest } Op.andOp)
  | Op.orOp =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opOr x y :: rest } Op.orOp)
  | Op.xorOp =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opXor x y :: rest } Op.xorOp)
  | Op.notOp =>
      match pop1 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, rest) =>
          Except.ok (withGas { s with stack := opNot x :: rest } Op.notOp)
  | Op.lt =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opLt x y :: rest } Op.lt)
  | Op.gt =>
      match pop2 s.stack with
      | none => Except.error Error.stackUnderflow
      | some (x, y, rest) =>
          Except.ok (withGas { s with stack := opGt x y :: rest } Op.gt)
  | Op.dup n =>
      if h : n = 0 then
        Except.error Error.stackUnderflow
      else
        match nthFromTop s.stack (n - 1) with
        | none => Except.error Error.stackUnderflow
        | some v =>
            Except.ok (withGas { s with stack := v :: s.stack } (Op.dup n))
  | Op.jump =>
      Except.error Error.invalidOperand
  | Op.jumpi =>
      Except.error Error.invalidOperand
  | Op.pop =>
      match popn 1 s.stack with
      | none => Except.error Error.stackUnderflow
      | some ([_], rest) =>
          Except.ok (withGas { s with stack := rest } Op.pop)
      | some (_, _) => Except.error Error.stackUnderflow
  | Op.jumpdest =>
      Except.ok (withGas s Op.jumpdest)
  | Op.swap n =>
      match swapTop s.stack n with
      | none => Except.error Error.stackUnderflow
      | some swapped =>
          Except.ok (withGas { s with stack := swapped } (Op.swap n))

def run (cfg : Config) (ops : Program) (s : State := {}) : Except Error State :=
  if ops = [] then
    Except.error Error.emptyProgram
  else
    let maxCycles := normalizeCycles cfg
    let fuel0 := maxCycles - s.steps
    let rec loop (fuel pc : Nat) (state : State) : Except Error State :=
      if fuel = 0 then
        Except.error Error.cycleLimit
        else if pc >= ops.length then
        Except.ok state
      else
        match listGet? ops pc with
        | none => Except.ok state
        | some Op.stop =>
            match exec state Op.stop with
            | Except.error e => Except.error e
            | Except.ok next =>
                Except.ok (withGas next Op.stop)
        | some Op.jump =>
            match pop1 state.stack with
            | none => Except.error Error.stackUnderflow
            | some (dest, rest) =>
                let nextState := withGas { state with stack := rest } Op.jump
                if hdest : dest < ops.length then
                  match listGet? ops dest with
                  | some Op.jumpdest =>
                      loop (fuel - 1) dest nextState
                  | _ =>
                      Except.error Error.invalidOperand
                else
                  Except.error Error.invalidOperand
        | some Op.jumpi =>
            match pop2 state.stack with
            | none => Except.error Error.stackUnderflow
            | some (cond, dest, rest) =>
                let nextState := withGas { state with stack := rest } Op.jumpi
                if cond = 0 then
                  loop (fuel - 1) (pc + 1) nextState
                else if hdest : dest < ops.length then
                  match listGet? ops dest with
                  | some Op.jumpdest =>
                      loop (fuel - 1) dest nextState
                  | _ =>
                      Except.error Error.invalidOperand
                else
                  Except.error Error.invalidOperand
        | some op =>
            match exec state op with
            | Except.error e => Except.error e
            | Except.ok next =>
                loop (fuel - 1) (pc + 1) next
    loop fuel0 0 s

end Lean2030.EVM
