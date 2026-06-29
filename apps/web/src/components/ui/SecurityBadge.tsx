export function SecurityBadge({ status, children, className }) {
  // Semantic colors from the design system
  const colorMap = {
    success: 'bg-green-600 text-white',
    warning: 'bg-yellow-600 text-white',
    danger:  'bg-red-600 text-white',
    info:    'bg-blue-600 text-white',
    mfa:     'bg-purple-600 text-white',
  };
  const bg = colorMap[status] || colorMap.success;
  const classes = `px-2 py-1 rounded text-xs font-medium whitespace-nowrap ${bg} ${className || ''}`;
  return <span className={classes}>{children}</span>;
}