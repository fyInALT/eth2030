import Lean2030.VM.Spec

namespace Lean2030.VM

inductive CompileError where
  | emptyProgram
  | invalidPush
  | unsupportedOpcode
  deriving Repr, BEq, DecidableEq

/-- Translate a tiny EVM bytecode subset into LeanVM ops.
    Supported opcodes:
    - 0x01 ADD
    - 0x02 MUL
    - 0x14 EQ
    - 0x20 SHA3
    - 0x60 PUSH1 (with one immediate)
    - 0x80 DUP1
-/
def compile (bytecode : List UInt8) : Except CompileError Program :=
  if bytecode = [] then
    Except.error CompileError.emptyProgram
  else
    let rec aux (bs : List UInt8) (inputIdx : Nat) : Except CompileError Program := do
      match bs with
      | [] => pure []
      | b :: rest =>
          match b with
          | 0x01 =>
              let ops ← aux rest inputIdx
              Except.ok (Opcode.add :: ops)
          | 0x02 =>
              let ops ← aux rest inputIdx
              Except.ok (Opcode.mul :: ops)
          | 0x14 =>
              let ops ← aux rest inputIdx
              Except.ok (Opcode.verify :: ops)
          | 0x20 =>
              let ops ← aux rest inputIdx
              Except.ok (Opcode.hash :: ops)
          | 0x60 =>
              match rest with
              | [] => Except.error CompileError.invalidPush
              | _b :: rest' =>
                  let ops ← aux rest' (inputIdx + 1)
                  Except.ok (Opcode.push inputIdx :: ops)
          | 0x80 =>
              let ops ← aux rest inputIdx
              Except.ok (Opcode.dup :: ops)
          | _ =>
              Except.error CompileError.unsupportedOpcode
    aux bytecode 0

end Lean2030.VM
