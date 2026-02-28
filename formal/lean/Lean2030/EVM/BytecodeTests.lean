import Lean2030.EVM.Bytecode
import Lean2030.EVM.Compile

namespace Lean2030.EVM

instance {ε α : Type} [DecidableEq ε] [DecidableEq α] (a b : Except ε α) :
    Decidable (a = b) := by
  cases a with
  | error e1 =>
    cases b with
    | error e2 =>
      match decEq e1 e2 with
      | isTrue h =>
          exact isTrue (by simpa [h])
      | isFalse h =>
          exact isFalse (by intro hEq; cases hEq; exact h rfl)
    | ok a2 =>
      exact isFalse (by intro hEq; cases hEq)
  | ok a1 =>
    cases b with
    | error e2 =>
      exact isFalse (by intro hEq; cases hEq)
    | ok a2 =>
      match decEq a1 a2 with
      | isTrue h =>
          exact isTrue (by simpa [h])
      | isFalse h =>
          exact isFalse (by intro hEq; cases hEq; exact h rfl)

section Decode

example : decodeNoImmediate 0x00 = some Op.stop := by native_decide

example : decodeNoImmediate 0x56 = some Op.jump := by native_decide

example : decodeNoImmediate 0x57 = some Op.jumpi := by native_decide

example : decodeNoImmediate 0x5f = some (Op.push 0) := by native_decide

example : decodeAt [0x60, 0x2a, 0x00] 0 = some (Op.push (toWord 42), 2) := by native_decide

example : decodeAt [0x56] 0 = some (Op.jump, 1) := by native_decide

end Decode

section RunBytecodeBasics

example : runBytecode { maxCycles := 10 } [0x5f, 0x00] {} =
    Except.ok { stack := [0], gasUsed := gasOf (Op.push 0) + gasOf Op.stop, steps := 2 } := by
  native_decide

example : runBytecode { maxCycles := 10 } [0x60, 0x2a, 0x00] {} =
    Except.ok { stack := [toWord 42], gasUsed := gasOf (Op.push (toWord 42)), steps := 2 } := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x03, 0x56, 0x5b, 0x00] {} =
    Except.ok
      { stack := []
      , gasUsed := gasOf (Op.push 3) + gasOf Op.jump + gasOf Op.jumpdest + gasOf Op.stop
      , steps := 4 } := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x02, 0x56, 0x5b, 0x00] {} = Except.error Error.invalidOperand := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x02, 0x00] {} = Except.ok { stack := [toWord 2], gasUsed := gasOf (Op.push 2), steps := 2 } := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x03, 0x60, 0x00, 0x57, 0x00, 0x00] {} =
    Except.ok { stack := [], gasUsed := gasOf (Op.push 3) + gasOf (Op.push 0) + gasOf Op.jumpi + gasOf Op.stop, steps := 4 } := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x02, 0x60, 0x01, 0x57, 0x00, 0x00] {} =
    Except.error Error.invalidOperand := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x06, 0x60, 0x01, 0x57, 0x00, 0x5b, 0x00] {} =
    Except.ok { stack := [], gasUsed := gasOf (Op.push 6) + gasOf (Op.push 1) + gasOf Op.jumpi + gasOf Op.jumpdest + gasOf Op.stop, steps := 5 } := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60] {} = Except.error Error.unsupported := by
  native_decide

example : runBytecode { maxCycles := 20 } [0x60, 0x10] {} =
    Except.ok { stack := [toWord 16], gasUsed := gasOf (Op.push 16), steps := 1 } := by
  native_decide

example : runBytecode { maxCycles := 1 } [0x60, 0x01, 0x00] {} = Except.error Error.cycleLimit := by
  native_decide

end RunBytecodeBasics

section CompileBytecodeBridge

example : compile [0x56] = Except.ok [Op.jump] := by
  native_decide

example : compile [0x57] = Except.ok [Op.jumpi] := by
  native_decide

example : compileNoImmediate [0x56] = some [Op.jump] := by
  native_decide

example : compileNoImmediate [0x57] = some [Op.jumpi] := by
  native_decide

example : compile [0x60, 0x01, 0x56, 0x5b, 0x00] =
    Except.ok [Op.push 1, Op.jump, Op.jumpdest, Op.stop] := by
  native_decide

example : compileNoImmediate [0x60, 0x01, 0x56, 0x5b, 0x00] = none := by
  native_decide

example : compileNoImmediate [0x60, 0x56] = none := by
  native_decide

end CompileBytecodeBridge

end Lean2030.EVM
