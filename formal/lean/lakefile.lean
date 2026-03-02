import Lake
open Lake DSL

package lean2030 where
  testDriver := "lean2030Tests"

lean_exe lean2030Tests where
  root := `Lean2030.Tests
  supportInterpreter := true

lean_lib Lean2030 where
