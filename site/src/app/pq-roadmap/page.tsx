'use client';
import { motion } from 'framer-motion';
import Link from 'next/link';
import SectionHeading from '@/components/ui/SectionHeading';
import { useInView } from '@/hooks/useInView';
import { GITHUB_URL } from '@/lib/constants';

const fadeIn = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const coverageData = [
  { item: 'Hash-based sigs (WOTS+/XMSS)', pkg: 'pkg/crypto/pqc/', status: 'complete' },
  { item: 'Pluggable hash functions', pkg: 'pkg/crypto/pqc/hash_backend.go', status: 'complete' },
  { item: 'STARK proof aggregation', pkg: 'pkg/proofs/stark_prover.go', status: 'complete' },
  { item: 'Recursive STARK composition', pkg: 'pkg/proofs/recursive_prover.go', status: 'complete' },
  { item: 'STARK mempool aggregation', pkg: 'pkg/txpool/stark_aggregation.go', status: 'complete' },
  { item: 'STARK CL sig aggregation', pkg: 'pkg/consensus/stark_sig_aggregation.go', status: 'complete' },
  { item: 'EIP-8141 frame transactions', pkg: 'pkg/core/ (17 files)', status: 'complete' },
  { item: 'NTT precompile (EIP-7885)', pkg: 'pkg/core/vm/precompile_ntt.go', status: 'complete' },
  { item: 'Lattice sigs (Dilithium/Falcon)', pkg: 'pkg/crypto/pqc/', status: 'complete' },
  { item: 'PQ attestations', pkg: 'pkg/consensus/pq_attestation.go', status: 'complete' },
  { item: 'Lattice blob commitments', pkg: 'pkg/crypto/pqc/lattice_blob.go', status: 'complete' },
  { item: 'PQ algorithm registry', pkg: 'pkg/crypto/pqc/registry.go', status: 'complete' },
];

const vulnerableAreas = [
  {
    id: 'cl-bls',
    title: 'CL BLS Signatures',
    threat: 'BLS12-381 signatures used for consensus attestations are vulnerable to Shor\'s algorithm.',
    solution: 'STARK-aggregated hash-based signatures. Validators sign with Dilithium3, then a single STARK proves all N signatures are valid.',
    packages: ['pkg/consensus/pq_attestation.go', 'pkg/consensus/stark_sig_aggregation.go', 'pkg/consensus/jeanvm_aggregation.go'],
    color: { border: 'border-eth-purple/40', bg: 'bg-eth-purple/10', text: 'text-eth-purple', glow: 'shadow-[0_0_10px_#8c8dfc22]' },
  },
  {
    id: 'da-kzg',
    title: 'DA KZG Commitments',
    threat: 'KZG polynomial commitments rely on elliptic curve pairings vulnerable to quantum attacks.',
    solution: 'Lattice-based blob commitments using Module-LWE (MLWE). Dual-commit (KZG + MLWE) during migration.',
    packages: ['pkg/crypto/pqc/lattice_blob.go', 'pkg/das/'],
    color: { border: 'border-eth-teal/40', bg: 'bg-eth-teal/10', text: 'text-eth-teal', glow: 'shadow-[0_0_10px_#2de2e622]' },
  },
  {
    id: 'eoa-ecdsa',
    title: 'EOA ECDSA Signatures',
    threat: 'ECDSA signatures on user transactions are vulnerable to Shor\'s algorithm.',
    solution: 'EIP-8141 frame transactions with PQ algorithm registry. Programmable tx validation with Dilithium, Falcon, SPHINCS+, WOTS+/XMSS.',
    packages: ['pkg/core/ (EIP-8141)', 'pkg/crypto/pqc/registry.go', 'pkg/crypto/pqc/hash_backend.go'],
    color: { border: 'border-eth-blue/40', bg: 'bg-eth-blue/10', text: 'text-eth-blue', glow: 'shadow-[0_0_10px_#627eea22]' },
  },
  {
    id: 'app-proofs',
    title: 'Application-layer Proofs',
    threat: 'ZK proofs verified on-chain may use quantum-vulnerable assumptions (elliptic curve pairings).',
    solution: 'Recursive STARK mempool aggregation. Every 500ms, nodes create a STARK proving validity of all validated transactions.',
    packages: ['pkg/proofs/stark_prover.go', 'pkg/txpool/stark_aggregation.go', 'pkg/core/vm/precompile_ntt.go'],
    color: { border: 'border-eth-pink/40', bg: 'bg-eth-pink/10', text: 'text-eth-pink', glow: 'shadow-[0_0_10px_#ff6b9d22]' },
  },
];

const flowSteps = {
  txFlow: [
    { label: 'User (PQ Wallet)', sub: 'Signs with Dilithium/Falcon' },
    { label: 'EIP-8141 Frame Tx', sub: 'Type 0x06' },
    { label: 'PQ Sig Verify', sub: 'Algorithm Registry' },
    { label: 'STARK Aggregator', sub: 'Every 500ms' },
    { label: 'Block Builder', sub: 'Aggregated proof' },
    { label: 'CL Attestation', sub: 'STARK-aggregated' },
  ],
  proofPipeline: [
    { label: 'Individual Tx Proofs', sub: 'Groth16 / STARK' },
    { label: 'Batch STARK', sub: 'FRI commitment' },
    { label: 'Recursive Composition', sub: 'Outer STARK' },
    { label: 'Block-level Proof', sub: 'Single verification' },
  ],
};

const refSubmodules = [
  { name: 'refs/hash-sig', repo: 'b-wagn/hash-sig', desc: 'Rust hash-based multi-signature library. Supports SHA3 and Poseidon2.' },
  { name: 'refs/ntt-eip', repo: 'ZKNoxHQ/NTT', desc: 'EIP-7885 NTT precompile reference implementation (Solidity + Python).' },
  { name: 'refs/ethfalcon', repo: 'ZKNoxHQ/ETHFALCON', desc: 'FALCON-512 signature verification on EVM with NTT optimization.' },
];

function FlowDiagram({ steps, title, color }: { steps: typeof flowSteps.txFlow; title: string; color: string }) {
  return (
    <div className="mb-10">
      <h4 className={`text-lg font-bold ${color} mb-4`}>{title}</h4>
      <div className="flex flex-wrap items-center gap-2">
        {steps.map((step, i) => (
          <div key={step.label} className="flex items-center gap-2">
            <div className="rounded-lg border border-eth-purple/20 bg-eth-surface p-3 text-center min-w-[140px]">
              <div className="text-sm font-semibold text-eth-text">{step.label}</div>
              <div className="text-xs text-eth-dim mt-1">{step.sub}</div>
            </div>
            {i < steps.length - 1 && (
              <span className="text-eth-purple text-lg font-bold">&rarr;</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export default function PQRoadmapPage() {
  const { ref: coverageRef, isInView: coverageInView } = useInView();
  const { ref: areasRef, isInView: areasInView } = useInView();
  const { ref: flowRef, isInView: flowInView } = useInView();

  return (
    <main className="min-h-screen">
      {/* Header */}
      <section className="relative py-20 md:py-28 px-4 text-center">
        <div className="absolute inset-0 bg-grid-pattern opacity-50" aria-hidden="true" />
        <div
          className="absolute inset-0"
          style={{ background: 'radial-gradient(ellipse at center, rgba(58, 28, 113, 0.2), transparent 70%)' }}
          aria-hidden="true"
        />
        <div className="relative z-10 max-w-4xl mx-auto">
          <motion.div {...{ initial: fadeIn.hidden, animate: fadeIn.visible }} transition={{ duration: 0.6 }}>
            <Link
              href="/"
              className="inline-block mb-6 px-4 py-1.5 rounded-full border border-eth-purple/30
                         bg-eth-purple/5 text-eth-purple text-sm tracking-wider uppercase hover:bg-eth-purple/10 transition-colors"
            >
              &larr; Back to ETH2030
            </Link>
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="text-4xl sm:text-5xl md:text-6xl font-bold tracking-tighter neon-purple"
          >
            Post-Quantum Roadmap
          </motion.h1>

          <motion.p
            initial={fadeIn.hidden}
            animate={fadeIn.visible}
            transition={{ duration: 0.6, delay: 0.5 }}
            className="mt-6 text-lg text-eth-dim max-w-2xl mx-auto leading-relaxed"
          >
            Aligning eth2030 with Vitalik&apos;s Ethereum quantum resistance roadmap.
            Covering <span className="text-eth-purple">4 vulnerable areas</span>,{' '}
            <span className="text-eth-teal">7 packages</span>, and a{' '}
            <span className="text-eth-pink">pluggable architecture</span> for hash function agility.
          </motion.p>
        </div>
      </section>

      {/* 4 Vulnerable Areas */}
      <section className="py-16 md:py-24 px-4 bg-eth-surface/30">
        <SectionHeading
          title="4 Vulnerable Areas"
          subtitle="Each area of Ethereum at risk from quantum attacks, and how eth2030 addresses it"
        />

        <div ref={areasRef} className="max-w-5xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-6">
          {vulnerableAreas.map((area, i) => (
            <motion.div
              key={area.id}
              initial={{ opacity: 0, y: 30 }}
              animate={areasInView ? { opacity: 1, y: 0 } : {}}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className={`rounded-xl border ${area.color.border} ${area.color.bg} ${area.color.glow} p-6`}
            >
              <h3 className={`text-xl font-bold ${area.color.text} mb-3`}>{area.title}</h3>
              <div className="mb-3">
                <span className="text-xs uppercase tracking-wider text-eth-dim">Threat:</span>
                <p className="text-sm text-eth-text/80 mt-1">{area.threat}</p>
              </div>
              <div className="mb-3">
                <span className="text-xs uppercase tracking-wider text-eth-dim">Solution:</span>
                <p className="text-sm text-eth-text/80 mt-1">{area.solution}</p>
              </div>
              <div>
                <span className="text-xs uppercase tracking-wider text-eth-dim">Packages:</span>
                <div className="flex flex-wrap gap-1 mt-1">
                  {area.packages.map((pkg) => (
                    <span key={pkg} className="text-xs px-2 py-0.5 rounded bg-eth-bg/50 text-eth-dim border border-eth-purple/10">
                      {pkg}
                    </span>
                  ))}
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Workflow Diagrams */}
      <section className="py-16 md:py-24 px-4">
        <SectionHeading
          title="Workflow"
          subtitle="How post-quantum transactions and proofs flow through eth2030"
        />

        <div ref={flowRef} className="max-w-5xl mx-auto">
          <motion.div
            initial={fadeIn.hidden}
            animate={flowInView ? fadeIn.visible : {}}
            transition={{ duration: 0.6 }}
          >
            <FlowDiagram steps={flowSteps.txFlow} title="PQ Transaction Flow" color="text-eth-purple" />
            <FlowDiagram steps={flowSteps.proofPipeline} title="Proof Aggregation Pipeline" color="text-eth-teal" />
          </motion.div>

          {/* Hash Backend Diagram */}
          <motion.div
            initial={fadeIn.hidden}
            animate={flowInView ? fadeIn.visible : {}}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="mt-4"
          >
            <h4 className="text-lg font-bold text-eth-pink mb-4">Hash Function Pluggability</h4>
            <div className="rounded-xl border border-eth-pink/20 bg-eth-surface p-6">
              <div className="flex flex-col md:flex-row items-center gap-6">
                <div className="rounded-lg border border-eth-purple/30 bg-eth-bg p-4 text-center">
                  <div className="text-sm font-bold text-eth-purple">HashBackend Interface</div>
                  <div className="text-xs text-eth-dim mt-1">Hash() / Name() / BlockSize()</div>
                </div>
                <span className="text-eth-purple text-lg font-bold hidden md:block">&rarr;</span>
                <span className="text-eth-purple text-lg font-bold md:hidden">&darr;</span>
                <div className="flex flex-wrap gap-2 justify-center">
                  {['Keccak256', 'SHA-256', 'BLAKE3', 'Poseidon2 (future)'].map((h) => (
                    <span key={h} className="px-3 py-1.5 rounded-lg border border-eth-teal/20 bg-eth-teal/5 text-sm text-eth-teal">
                      {h}
                    </span>
                  ))}
                </div>
                <span className="text-eth-purple text-lg font-bold hidden md:block">&rarr;</span>
                <span className="text-eth-purple text-lg font-bold md:hidden">&darr;</span>
                <div className="flex flex-wrap gap-2 justify-center">
                  {['WOTS+', 'XMSS', 'L1 Signer'].map((s) => (
                    <span key={s} className="px-3 py-1.5 rounded-lg border border-eth-pink/20 bg-eth-pink/5 text-sm text-eth-pink">
                      {s}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Coverage Matrix */}
      <section className="py-16 md:py-24 px-4 bg-eth-surface/30">
        <SectionHeading
          title="Coverage Matrix"
          subtitle="Mapping Vitalik's roadmap items to eth2030 implementations"
        />

        <div ref={coverageRef} className="max-w-5xl mx-auto overflow-x-auto">
          <motion.table
            initial={fadeIn.hidden}
            animate={coverageInView ? fadeIn.visible : {}}
            transition={{ duration: 0.6 }}
            className="w-full text-sm"
          >
            <thead>
              <tr className="border-b border-eth-purple/20">
                <th className="text-left py-3 px-4 text-eth-purple font-semibold">Roadmap Item</th>
                <th className="text-left py-3 px-4 text-eth-purple font-semibold">Package</th>
                <th className="text-center py-3 px-4 text-eth-purple font-semibold">Status</th>
              </tr>
            </thead>
            <tbody>
              {coverageData.map((row, i) => (
                <motion.tr
                  key={row.item}
                  initial={{ opacity: 0, x: -10 }}
                  animate={coverageInView ? { opacity: 1, x: 0 } : {}}
                  transition={{ duration: 0.3, delay: i * 0.05 }}
                  className="border-b border-eth-purple/10 hover:bg-eth-purple/5 transition-colors"
                >
                  <td className="py-2.5 px-4 text-eth-text">{row.item}</td>
                  <td className="py-2.5 px-4 text-eth-dim font-mono text-xs">{row.pkg}</td>
                  <td className="py-2.5 px-4 text-center">
                    <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold bg-eth-teal/10 text-eth-teal border border-eth-teal/20">
                      {row.status}
                    </span>
                  </td>
                </motion.tr>
              ))}
            </tbody>
          </motion.table>
        </div>
      </section>

      {/* Reference Implementations */}
      <section className="py-16 md:py-24 px-4">
        <SectionHeading
          title="References"
          subtitle="Reference submodules supporting the PQ roadmap"
        />

        <div className="max-w-5xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-6">
          {refSubmodules.map((ref, i) => (
            <motion.div
              key={ref.name}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="rounded-xl border border-eth-purple/20 bg-eth-surface p-6"
            >
              <h4 className="text-eth-purple font-bold mb-1">{ref.name}</h4>
              <a
                href={`https://github.com/${ref.repo}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-eth-teal hover:underline"
              >
                {ref.repo}
              </a>
              <p className="text-sm text-eth-dim mt-3">{ref.desc}</p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* NTT Precompile */}
      <section className="py-16 md:py-24 px-4 bg-eth-surface/30">
        <SectionHeading
          title="NTT Precompile"
          subtitle="EIP-7885 Number Theoretic Transform at address 0x15"
        />

        <div className="max-w-3xl mx-auto">
          <div className="rounded-xl border border-eth-purple/20 bg-eth-bg p-6">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-eth-purple/20">
                  <th className="text-left py-2 px-3 text-eth-purple">Op</th>
                  <th className="text-left py-2 px-3 text-eth-purple">Field</th>
                  <th className="text-left py-2 px-3 text-eth-purple">Use Case</th>
                </tr>
              </thead>
              <tbody className="text-eth-dim">
                <tr className="border-b border-eth-purple/10">
                  <td className="py-2 px-3 font-mono text-eth-teal">0x00</td>
                  <td className="py-2 px-3">BN254 (254-bit)</td>
                  <td className="py-2 px-3">ZK-SNARK circuits, Groth16</td>
                </tr>
                <tr className="border-b border-eth-purple/10">
                  <td className="py-2 px-3 font-mono text-eth-teal">0x01</td>
                  <td className="py-2 px-3">BN254 (inverse)</td>
                  <td className="py-2 px-3">Polynomial interpolation</td>
                </tr>
                <tr className="border-b border-eth-purple/10">
                  <td className="py-2 px-3 font-mono text-eth-pink">0x02</td>
                  <td className="py-2 px-3">Goldilocks (2^64-2^32+1)</td>
                  <td className="py-2 px-3">STARK proofs, FRI, Plonky2</td>
                </tr>
                <tr>
                  <td className="py-2 px-3 font-mono text-eth-pink">0x03</td>
                  <td className="py-2 px-3">Goldilocks (inverse)</td>
                  <td className="py-2 px-3">STARK polynomial recovery</td>
                </tr>
              </tbody>
            </table>
            <p className="text-xs text-eth-dim mt-4">
              Gas cost: base(1000) + n &times; log2(n) &times; 10
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <section className="py-12 px-4 text-center">
        <div className="max-w-2xl mx-auto">
          <p className="text-eth-dim text-sm">
            Full report available at{' '}
            <a
              href={`${GITHUB_URL}/blob/master/docs/PQ_ROADMAP_REPORT.md`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-eth-purple hover:underline"
            >
              docs/PQ_ROADMAP_REPORT.md
            </a>
          </p>
          <div className="mt-6">
            <Link
              href="/"
              className="px-6 py-2 rounded-lg bg-eth-purple/10 border border-eth-purple/50
                         text-eth-purple font-semibold hover:bg-eth-purple/20
                         hover:shadow-[0_0_20px_#8c8dfc33] transition-all duration-300"
            >
              &larr; Back to ETH2030
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}
