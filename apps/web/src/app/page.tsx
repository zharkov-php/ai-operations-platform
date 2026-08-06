import Link from "next/link";

export default function Home() {
  return (
    <main className="page-shell">
      <p className="eyebrow">AI Execution Advisor</p>
      <h1>Make AI execution decisions with evidence.</h1>
      <p className="lede">Observe workload cost, evaluate safer alternatives, and distinguish estimated opportunities from verified savings.</p>
      <Link className="primary-link" href="/dashboard">View platform foundation</Link>
    </main>
  );
}
