import Lean2030.VM.Spec
import Lean2030.VM.Compile

namespace Lean2030.VM

/-- Hash bytewise XOR is length-maximal.
-/
theorem addBytes_len (a b : Value) : (addBytes a b).length = Nat.max a.length b.length := by
  induction a generalizing b with
  | nil =>
      cases b <;> simp [addBytes]
  | cons x xs ih =>
      cases b with
      | nil => simp [addBytes, ih]
      | cons y ys =>
          simp [addBytes, ih]

/-- A successful PUSH appends one value and charges PUSH gas.
-/
theorem execOp_push_success
  (inputs : InputTape)
  (idx : Nat)
  (v : Value)
  (h : listGet? inputs idx = some v)
  (s : RunState) :
  execOp defaultHash inputs s (Opcode.push idx) =
    Except.ok { s with
      stack := v :: s.stack
      gasUsed := s.gasUsed + gasPush
      steps := s.steps } := by
  simp [execOp, h]

/-- VERIFY pushes a one-byte truth value based on argument equality.
-/
theorem execOp_verify_eq
  (a b : Value) (inputs : InputTape) (s : RunState) :
  execOp defaultHash inputs
    (s := { s with stack := a :: b :: s.stack })
    Opcode.verify =
    Except.ok { s with
      stack := [ [if a == b then (1 : UInt8) else 0] ] ++ s.stack
      gasUsed := s.gasUsed + gasVerify
      steps := s.steps } := by
  by_cases h : a == b
  · simp [execOp, pop2, h]
  · simp [execOp, pop2, h]

/-- compile rejects empty bytecode.
-/
theorem compile_empty_error : compile ([] : List UInt8) = Except.error CompileError.emptyProgram := by
  simp [compile]

/-- A single-byte nonempty PUSH1 bytecode fails only if the push operand is absent.
    Two-byte `PUSH1 <x>` at least compiles to a single push instruction.
-/
theorem compile_push1_two_bytes
  (x : UInt8) : compile [0x60, x] = Except.ok [Opcode.push 0] := by
  simp [compile]

/-- `run` reports empty-program error exactly.
-/
theorem run_empty_error (inputs : InputTape) (cfg : Config) :
  run defaultHash inputs cfg [] = Except.error ExecError.T.emptyProgram := by
  simp [run]

/-- A single PUSH instruction executes to one stack element and one step.
-/
theorem run_single_push
  (inputs : InputTape)
  (idx : Nat)
  (v : Value)
  (h : listGet? inputs idx = some v) :
  run defaultHash inputs { maxCycles := 10 } [Opcode.push idx] =
    Except.ok
      (some v,
       { stack := [v], gasUsed := gasPush, steps := 1 }) := by
  simp [run, normalize, execOp, h]

end Lean2030.VM
