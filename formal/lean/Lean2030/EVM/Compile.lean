import Lean2030.EVM.Spec

namespace Lean2030.EVM

inductive CompileError where
  | emptyProgram
  | truncatedPush
  | unsupported
  deriving BEq, Repr, DecidableEq

def pushSize (b : UInt8) : Nat :=
  let n := b.toNat
  if h : 0x60 ≤ n ∧ n ≤ 0x7f then n - 0x60 + 1 else 0

def byteToNat : UInt8 → Nat := fun b => b.toNat

def bytesToWord : List UInt8 → Word
  | [] => 0
  | bs => bs.foldl (fun acc b => acc * 256 + byteToNat b) 0

def dupIndex (op : UInt8) : Option Nat :=
  let n := op.toNat
  if h : 0x80 ≤ n ∧ n ≤ 0x8f then some (Nat.succ (n - 0x80)) else none

def swapIndex (op : UInt8) : Option Nat :=
  let n := op.toNat
  if h : 0x90 ≤ n ∧ n ≤ 0x9f then some (n - 0x90 + 1) else none

/-- Compile bytecode into EVM ops.

    The compiler is implemented as a single-pass decoder so push immediates are consumed by
    tracking a pending byte budget. This avoids complicated structural recursion over
    non-tail-subevents in older Lean versions.
-/
def compile (bytecode : List UInt8) : Except CompileError Program :=
  if bytecode = [] then
    Except.error CompileError.emptyProgram
  else
    let rec loop (bs : List UInt8) (acc : Program) (pending : Nat) (accImm : Nat) : Except CompileError Program :=
      match bs with
      | [] =>
          if h : pending = 0 then
            Except.ok acc.reverse
          else
            Except.error CompileError.truncatedPush
      | b :: rest =>
          if hPending : pending ≠ 0 then
            let accImm' := accImm * 256 + byteToNat b
            let pending' := pending - 1
            if hdone : pending' = 0 then
              loop rest (Op.push (toWord accImm') :: acc) 0 0
            else
              loop rest acc pending' accImm'
          else
            match b with
            | 0x00 => loop rest (Op.stop :: acc) 0 0
            | 0x01 => loop rest (Op.add :: acc) 0 0
            | 0x02 => loop rest (Op.mul :: acc) 0 0
            | 0x03 => loop rest (Op.sub :: acc) 0 0
            | 0x04 => loop rest (Op.div :: acc) 0 0
            | 0x06 => loop rest (Op.mod :: acc) 0 0
            | 0x14 => loop rest (Op.eq :: acc) 0 0
            | 0x15 => loop rest (Op.isZero :: acc) 0 0
            | 0x16 => loop rest (Op.andOp :: acc) 0 0
            | 0x17 => loop rest (Op.orOp :: acc) 0 0
            | 0x18 => loop rest (Op.xorOp :: acc) 0 0
            | 0x19 => loop rest (Op.notOp :: acc) 0 0
            | 0x1a => Except.error CompileError.unsupported
            | 0x1b => Except.error CompileError.unsupported
            | 0x1c => Except.error CompileError.unsupported
            | 0x1d => Except.error CompileError.unsupported
            | 0x1e => Except.error CompileError.unsupported
            | 0x10 => loop rest (Op.lt :: acc) 0 0
            | 0x11 => loop rest (Op.gt :: acc) 0 0
            | 0x5f => loop rest (Op.push 0 :: acc) 0 0
            | 0x50 => loop rest (Op.pop :: acc) 0 0
            | 0x56 => loop rest (Op.jump :: acc) 0 0
            | 0x57 => loop rest (Op.jumpi :: acc) 0 0
            | 0x5b => loop rest (Op.jumpdest :: acc) 0 0
            | 0x60 => loop rest (acc) (pushSize 0x60) 0
            | 0x61 => loop rest (acc) (pushSize 0x61) 0
            | 0x62 => loop rest (acc) (pushSize 0x62) 0
            | 0x63 => loop rest (acc) (pushSize 0x63) 0
            | 0x64 => loop rest (acc) (pushSize 0x64) 0
            | 0x65 => loop rest (acc) (pushSize 0x65) 0
            | 0x66 => loop rest (acc) (pushSize 0x66) 0
            | 0x67 => loop rest (acc) (pushSize 0x67) 0
            | 0x68 => loop rest (acc) (pushSize 0x68) 0
            | 0x69 => loop rest (acc) (pushSize 0x69) 0
            | 0x6a => loop rest (acc) (pushSize 0x6a) 0
            | 0x6b => loop rest (acc) (pushSize 0x6b) 0
            | 0x6c => loop rest (acc) (pushSize 0x6c) 0
            | 0x6d => loop rest (acc) (pushSize 0x6d) 0
            | 0x6e => loop rest (acc) (pushSize 0x6e) 0
            | 0x6f => loop rest (acc) (pushSize 0x6f) 0
            | 0x70 => loop rest (acc) (pushSize 0x70) 0
            | 0x71 => loop rest (acc) (pushSize 0x71) 0
            | 0x72 => loop rest (acc) (pushSize 0x72) 0
            | 0x73 => loop rest (acc) (pushSize 0x73) 0
            | 0x74 => loop rest (acc) (pushSize 0x74) 0
            | 0x75 => loop rest (acc) (pushSize 0x75) 0
            | 0x76 => loop rest (acc) (pushSize 0x76) 0
            | 0x77 => loop rest (acc) (pushSize 0x77) 0
            | 0x78 => loop rest (acc) (pushSize 0x78) 0
            | 0x79 => loop rest (acc) (pushSize 0x79) 0
            | 0x7a => loop rest (acc) (pushSize 0x7a) 0
            | 0x7b => loop rest (acc) (pushSize 0x7b) 0
            | 0x7c => loop rest (acc) (pushSize 0x7c) 0
            | 0x7d => loop rest (acc) (pushSize 0x7d) 0
            | 0x7e => loop rest (acc) (pushSize 0x7e) 0
            | 0x7f => loop rest (acc) (pushSize 0x7f) 0
            | b =>
                if hPush : pushSize b > 0 then
                  loop rest acc (pushSize b) 0
                else
                  match dupIndex b with
                  | some n => loop rest (Op.dup n :: acc) 0 0
                  | none =>
                      match swapIndex b with
                      | some n => loop rest (Op.swap n :: acc) 0 0
                      | none =>
                          match b with
                          | 0xfe => loop rest (Op.stop :: acc) 0 0
                          | _ => Except.error CompileError.unsupported
    loop bytecode [] 0 0

end Lean2030.EVM
