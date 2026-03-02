import Lean2030.EVM.Spec

namespace Lean2030.EVM

theorem withGas_steps_eq (s : State) (op : Op) :
  (withGas s op).steps = s.steps + 1 := by
  rfl

theorem withGas_gas_eq (s : State) (op : Op) :
  (withGas s op).gasUsed = s.gasUsed + gasOf op := by
  rfl

theorem withGas_stack_eq (s : State) (op : Op) :
  (withGas s op).stack = s.stack := by
  rfl

theorem exec_steps_or_error (s : State) (op : Op) :
  match exec s op with
  | Except.ok s' => s'.steps = s.steps + 1
  | Except.error _ => True := by
  cases op <;> simp [exec, pop2, pop1, popn]

theorem exec_gas_or_error (s : State) (op : Op) :
  match exec s op with
  | Except.ok s' => s'.gasUsed = s.gasUsed + gasOf op
  | Except.error _ => True := by
  cases op <;> simp [exec, pop2, pop1, popn]

end Lean2030.EVM
