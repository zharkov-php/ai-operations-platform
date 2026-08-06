export function LoadingState({ label = "Loading platform data…" }: { label?: string }) { return <div className="state-card" aria-live="polite">{label}</div>; }
export function ErrorState({ message }: { message: string }) { return <div className="state-card state-error" role="alert"><strong>Data unavailable</strong><p>{message}</p></div>; }
export function EmptyState({ title, message }: { title: string; message: string }) { return <div className="state-card"><strong>{title}</strong><p>{message}</p></div>; }
export function DashboardHeading({ eyebrow, title, description }: { eyebrow: string; title: string; description: string }) { return <header className="dashboard-heading"><p className="eyebrow">{eyebrow}</p><h1>{title}</h1><p>{description}</p></header>; }
