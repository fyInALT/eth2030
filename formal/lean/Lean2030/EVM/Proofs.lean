import Lean2030.EVM.Spec
import Lean2030.EVM.Compile

namespace Lean2030.EVM

theorem popn_none_iff_too_short (xs : List α) :
  popn 2 xs = none ↔ xs.length < 2 := by
  cases xs <;> simp [popn]
  case cons x xs ih =>
    cases xs <;> simp [popn, ih]

theorem exec_push_increments_state (s : State) (v : Word) :
  exec s (Op.push v) = Except.ok (withGas (State.stack := v :: s.stack) s Op.push) := by
  simp [exec]

theorem exec_add_success (a b : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: b :: rest, gasUsed := g, steps := t } Op.add =
    Except.ok { stack := opAdd a b :: rest, gasUsed := g + gasOf Op.add, steps := t + 1 } := by
  simp [exec, popn]

theorem exec_add_underflow :
  exec (s := { stack := [1], gasUsed := 0, steps := 0 }) Op.add = Except.error Error.stackUnderflow := by
  simp [exec, popn]

theorem exec_mul_success (a b : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: b :: rest, gasUsed := g, steps := t } Op.mul =
    Except.ok { stack := opMul a b :: rest, gasUsed := g + gasOf Op.mul, steps := t + 1 } := by
  simp [exec, popn]

theorem exec_sub_success (a b : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: b :: rest, gasUsed := g, steps := t } Op.sub =
    Except.ok { stack := opSub a b :: rest, gasUsed := g + gasOf Op.sub, steps := t + 1 } := by
  simp [exec, popn]

theorem exec_not_success (a : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: rest, gasUsed := g, steps := t } Op.notOp =
    Except.ok { stack := opNot a :: rest, gasUsed := g + gasOf Op.notOp, steps := t + 1 } := by
  simp [exec, popn]

theorem exec_swap_1_success
  (a b : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: b :: rest, gasUsed := g, steps := t } (Op.swap 1) =
    Except.ok { stack := b :: a :: rest, gasUsed := g + gasOf (Op.swap 1), steps := t + 1 } := by
  simp [exec, swapTop, setNthFromTop]

theorem exec_dup_1_success (a : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: rest, gasUsed := g, steps := t } (Op.dup 1) =
    Except.ok { stack := a :: a :: rest, gasUsed := g + gasOf (Op.dup 1), steps := t + 1 } := by
  simp [exec]

theorem exec_dup_zero_underflow :
  exec { stack := [1, 2], gasUsed := 0, steps := 0 } (Op.dup 0) = Except.error Error.stackUnderflow := by
  simp [exec]

theorem exec_pop_success (a : Word) (rest : Stack) (g t : Nat) :
  exec { stack := a :: rest, gasUsed := g, steps := t } Op.pop =
    Except.ok { stack := rest, gasUsed := g + gasOf Op.pop, steps := t + 1 } := by
  simp [exec, popn]

theorem exec_jumpdest_success (rest : Stack) (g t : Nat) :
  exec { stack := rest, gasUsed := g, steps := t } Op.jumpdest =
    Except.ok { stack := rest, gasUsed := g + gasOf Op.jumpdest, steps := t + 1 } := by
  simp [exec]

theorem exec_jump_error (s : State) :
  exec s Op.jump = Except.error Error.invalidOperand := by
  simp [exec]

theorem exec_jumpi_error (s : State) :
  exec s Op.jumpi = Except.error Error.invalidOperand := by
  simp [exec]

theorem exec_stop_is_noop :
  exec (s := { stack := [1, 2], gasUsed := 7, steps := 9 }) Op.stop =
    Except.ok { stack := [1, 2], gasUsed := 7, steps := 9 } := by
  simp [exec]

theorem run_empty_error (cfg : Config) (s : State := {}) :
  run cfg [] s = Except.error Error.emptyProgram := by
  simp [run]

theorem run_single_push (cfg : Config) (v : Word) (g : Nat) :
  run { cfg with maxCycles := 10 } [Op.push v] { gasUsed := g, steps := 0, stack := [] } =
    Except.ok { stack := [v], gasUsed := g + gasOf (Op.push v), steps := 1 } := by
  simp [run, exec]

theorem run_stop_early (cfg : Config) :
  run { cfg with maxCycles := 10 } [Op.push 10, Op.stop, Op.mul] {} =
    Except.ok { stack := [10], gasUsed := gasOf (Op.push 10), steps := 2 } := by
  simp [run, exec]

theorem run_cycle_limit_hit :
  run { maxCycles := 1 } [Op.push 1, Op.push 2] {} =
    Except.error Error.cycleLimit := by
  simp [run, exec]

theorem run_jump_invalid_operand :
  run { maxCycles := 20 } [Op.push 5, Op.jump] {} =
    Except.error Error.invalidOperand := by
  simp [run, exec]

theorem run_jump_to_jumpdest :
  run { maxCycles := 20 } [Op.push 2, Op.jump, Op.jumpdest, Op.push 10] {} =
    Except.ok { stack := [10], gasUsed := gasOf (Op.push 2) + gasOf Op.jump + gasOf Op.jumpdest + gasOf (Op.push 10), steps := 4 } := by
  simp [run, exec, popn]

theorem run_jumpi_not_taken :
  run { maxCycles := 20 } [Op.push 2, Op.push 0, Op.jumpi, Op.push 9] {} =
    Except.ok { stack := [9], gasUsed := gasOf (Op.push 2) + gasOf (Op.push 0) + gasOf Op.jumpi + gasOf (Op.push 9), steps := 4 } := by
  simp [run, exec, popn]

theorem run_jumpi_taken :
  run { maxCycles := 20 } [Op.push 4, Op.push 1, Op.jumpi, Op.push 7, Op.jumpdest, Op.push 11] {} =
    Except.ok { stack := [11], gasUsed := gasOf (Op.push 4) + gasOf (Op.push 1) + gasOf Op.jumpi + gasOf Op.jumpdest + gasOf (Op.push 11), steps := 5 } := by
  simp [run, exec, popn]

theorem run_jumpi_invalid_jumpdest :
  run { maxCycles := 20 } [Op.push 3, Op.push 1, Op.jumpi, Op.push 9] {} =
    Except.error Error.invalidOperand := by
  simp [run, exec, popn]

theorem run_pop_underflow :
  run { maxCycles := 20 } [Op.pop] {} =
    Except.error Error.stackUnderflow := by
  simp [run, exec, popn]

theorem run_mul_underflow :
  run { maxCycles := 20 } [Op.mul] {} =
    Except.error Error.stackUnderflow := by
  simp [run, exec, popn]

theorem run_three_ops_push_add (cfg : Config) :
  run { cfg with maxCycles := 10 } [Op.push 2, Op.push 3, Op.add] {} =
    Except.ok { stack := [opAdd 3 2], gasUsed := gasOf (Op.push 2) + gasOf (Op.push 3) + gasOf Op.add, steps := 3 } := by
  simp [run, exec, popn]

theorem compile_push_size_zero_of_non_push (b : UInt8) (h : ¬ (0x60 ≤ b.toNat ∧ b.toNat ≤ 0x7f)) :
  pushSize b = 0 := by
  simp [pushSize, h]

theorem compile_push_size_eq (x : UInt8) (h : 0x60 ≤ x.toNat ∧ x.toNat ≤ 0x7f) :
  pushSize x = x.toNat - 0x60 + 1 := by
  simp [pushSize, h]

theorem dupIndex_eq_of_range (b : UInt8) (h : 0x80 ≤ b.toNat ∧ b.toNat ≤ 0x8f) :
  dupIndex b = some (Nat.succ (b.toNat - 0x80)) := by
  simp [dupIndex, h]

theorem swapIndex_eq_of_range (b : UInt8) (h : 0x90 ≤ b.toNat ∧ b.toNat ≤ 0x9f) :
  swapIndex b = some (b.toNat - 0x90 + 1) := by
  simp [swapIndex, h]

theorem dupIndex_none_of_not_range (b : UInt8) (h : ¬ (0x80 ≤ b.toNat ∧ b.toNat ≤ 0x8f)) :
  dupIndex b = none := by
  simp [dupIndex, h]

theorem swapIndex_none_of_not_range (b : UInt8) (h : ¬ (0x90 ≤ b.toNat ∧ b.toNat ≤ 0x9f)) :
  swapIndex b = none := by
  simp [swapIndex, h]

theorem bytesToWord_nil :
  bytesToWord ([] : List UInt8) = 0 := by
  simp [bytesToWord]

theorem bytesToWord_singleton (x : UInt8) :
  bytesToWord [x] = x.toNat := by
  simp [bytesToWord]

theorem bytesToWord_pair (x y : UInt8) :
  bytesToWord [x, y] = x.toNat * 256 + y.toNat := by
  simp [bytesToWord]

theorem compile_empty_error :
  compile ([] : List UInt8) = Except.error CompileError.emptyProgram := by
  simp [compile]

theorem compile_stop_error :
  compile [0xfe, 0x01] = Except.ok [Op.stop, Op.add] := by
  simp [compile]

theorem compile_push1_ok (x : UInt8) :
  compile [0x60, x] = Except.ok [Op.push (toWord x.toNat)] := by
  simp [compile, bytesToWord]

theorem compile_push0 :
  compile [0x5f] = Except.ok [Op.push 0] := by
  simp [compile]

theorem compile_push_truncated :
  compile [0x60] = Except.error CompileError.truncatedPush := by
  simp [compile, pushSize]

theorem compile_dup1 :
  compile [0x80] = Except.ok [Op.dup 1] := by
  simp [compile]

theorem compile_swap1 :
  compile [0x90] = Except.ok [Op.swap 1] := by
  simp [compile]

theorem compile_pop :
  compile [0x50] = Except.ok [Op.pop] := by
  simp [compile]

theorem compile_jump :
  compile [0x56] = Except.ok [Op.jump] := by
  simp [compile]

theorem compile_jumpi :
  compile [0x57] = Except.ok [Op.jumpi] := by
  simp [compile]

theorem compile_jumpdest :
  compile [0x5b] = Except.ok [Op.jumpdest] := by
  simp [compile]

theorem compile_add :
  compile [0x01] = Except.ok [Op.add] := by
  simp [compile]

theorem compile_push_add :
  compile [0x60, 0x02, 0x60, 0x03, 0x01] =
    Except.ok [Op.push (toWord 2), Op.push (toWord 3), Op.add] := by
  simp [compile, bytesToWord]

theorem compile_push2_ok :
  compile [0x61, 0x01, 0x02] =
    Except.ok [Op.push (toWord (1 * 256 + 2))] := by
  simp [compile, bytesToWord]

theorem compile_push2_truncated :
  compile [0x61, 0x01] = Except.error CompileError.truncatedPush := by
  simp [compile, pushSize]

theorem compile_push32_ok :
  compile [0x7f, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0] =
    Except.ok [Op.push 0] := by
  simp [compile, bytesToWord]

theorem compile_unsupported :
  compile [0xff] = Except.error CompileError.unsupported := by
  simp [compile]

theorem compile_truncated_push1 :
  compile [0x60] = Except.error CompileError.truncatedPush := by
  simp [compile, pushSize]

end Lean2030.EVM
