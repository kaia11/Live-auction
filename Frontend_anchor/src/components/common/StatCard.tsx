interface StatCardProps {
  label: string
  value: string
  tone?: 'normal' | 'success' | 'warning'
}

function StatCard({ label, value, tone = 'normal' }: StatCardProps) {
  return (
    <div className={`stat-card stat-card-${tone}`}>
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
    </div>
  )
}

export default StatCard
