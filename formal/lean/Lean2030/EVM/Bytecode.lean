import Lean2030.EVM.Spec
import Lean2030.EVM.Compile

namespace Lean2030.EVM

abbrev Bytecode := List UInt8

def byteAt? (bytecode : Bytecode) (pc : Nat) : Option UInt8 :=
  match bytecode with
  | [] => none
  | b :: bs =>
      match pc with
      | 0 => some b
      | pc + 1 => byteAt? bs pc

def decodeNoImmediate : UInt8 → Option Op
  | 0x00 => some Op.stop
  | 0x01 => some Op.add
  | 0x02 => some Op.mul
  | 0x03 => some Op.sub
  | 0x04 => some Op.div
  | 0x06 => some Op.mod
  | 0x14 => some Op.eq
  | 0x15 => some Op.isZero
  | 0x16 => some Op.andOp
  | 0x17 => some Op.orOp
  | 0x18 => some Op.xorOp
  | 0x19 => some Op.notOp
  | 0x10 => some Op.lt
  | 0x11 => some Op.gt
  | 0x50 => some Op.pop
  | 0x56 => some Op.jump
  | 0x57 => some Op.jumpi
  | 0x5b => some Op.jumpdest
  | 0x5f => some (Op.push 0)
  | b =>
      match dupIndex b with
      | some n => some (Op.dup n)
      | none =>
          match swapIndex b with
          | some n => some (Op.swap n)
          | none =>
              if b = 0xfe then
                some Op.stop
              else
                none

def compileNoImmediate (bytecode : Bytecode) : Option Program :=
  let rec loop : Bytecode → Option Program
    | [] => some []
    | b :: bs =>
        match decodeNoImmediate b, loop bs with
        | some op, some rest => some (op :: rest)
        | _, _ => none
  loop bytecode

def decodePush (bytecode : Bytecode) (pc n : Nat) : Option (Op × Nat) :=
  let imm := (bytecode.drop (pc + 1)).take n
  if hlen : imm.length = n then
    some (Op.push (toWord (bytesToWord imm)), pc + 1 + n)
  else
    none

def decodeAt (bytecode : Bytecode) (pc : Nat) : Option (Op × Nat) :=
  match byteAt? bytecode pc with
  | none => none
  | some 0x5f =>
      some (Op.push 0, pc + 1)
  | some b =>
      let n := pushSize b
      if h : n > 0 then
        decodePush bytecode pc n
      else
        match decodeNoImmediate b with
        | none => none
        | some op => some (op, pc + 1)

/-- Execute bytecode directly at byte-offset `pc`, with EVM-style gas/step
    semantics and jump destination checks against raw byte offsets. -/
def runBytecode (cfg : Config) (bytecode : Bytecode) (s : State := {}) : Except Error State :=
  if bytecode = [] then
    Except.error Error.emptyProgram
  else
    let maxCycles := normalizeCycles cfg
    let rec loop (pc : Nat) (fuel : Nat) (state : State) : Except Error State :=
      if fuel = 0 then
        Except.error Error.cycleLimit
      else if pc >= bytecode.length then
        Except.ok state
      else
        match decodeAt bytecode pc with
        | none =>
            Except.error Error.unsupported
        | some (Op.stop, _) =>
            match exec state Op.stop with
            | Except.error e => Except.error e
            | Except.ok next =>
                Except.ok (withGas next Op.stop)
        | some (Op.jump, _) =>
            match popn 1 state.stack with
            | none => Except.error Error.stackUnderflow
            | some ([dest], rest) =>
                let nextState := withGas { state with stack := rest } Op.jump
                if byteAt? bytecode dest = some 0x5b then
                  loop dest (fuel - 1) nextState
                else
                  Except.error Error.invalidOperand
            | some (_, _) => Except.error Error.stackUnderflow
        | some (Op.jumpi, _) =>
            match popn 2 state.stack with
            | none => Except.error Error.stackUnderflow
            | some ([cond, dest], rest) =>
                let nextState := withGas { state with stack := rest } Op.jumpi
                if cond = 0 then
                  loop (pc + 1) (fuel - 1) nextState
                else if byteAt? bytecode dest = some 0x5b then
                  loop dest (fuel - 1) nextState
                else
                  Except.error Error.invalidOperand
            | some (_, _) => Except.error Error.stackUnderflow
        | some (op, nextPc) =>
            match exec state op with
            | Except.error e => Except.error e
            | Except.ok next =>
                loop nextPc (fuel - 1) next
    loop 0 (maxCycles - s.steps) s

theorem runBytecode_empty_error (cfg : Config) (s : State := {}) :
  runBytecode cfg [] s = Except.error Error.emptyProgram := by
  simp [runBytecode]


end Lean2030.EVM
