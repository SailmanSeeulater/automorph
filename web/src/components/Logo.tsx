// A sail in a circle. Uses currentColor, so it follows the text colour (and dark mode).
export function Logo({ size = 40 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 48 48"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="24" cy="24" r="22" />
      <path d="M24 10v24M24 12l10 20H24M14 36h20" />
    </svg>
  )
}
