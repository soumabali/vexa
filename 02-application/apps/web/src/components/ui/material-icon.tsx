import { cn } from "@/lib/utils";

interface MaterialIconProps {
  name: string;
  className?: string;
  fill?: boolean;
  weight?: 100 | 200 | 300 | 400 | 500 | 600 | 700;
  size?: "sm" | "md" | "lg" | "xl";
}

const sizeMap = {
  sm: "text-[14px]",
  md: "text-[18px]",
  lg: "text-[20px]",
  xl: "text-[24px]",
};

export function MaterialIcon({
  name,
  className,
  fill = false,
  weight = 400,
  size = "md",
}: MaterialIconProps) {
  return (
    <span
      className={cn("material-symbols-outlined", sizeMap[size], className)}
      style={{
        fontVariationSettings: `'FILL' ${fill ? 1 : 0}, 'wght' ${weight}, 'GRAD' 0, 'opsz' 24`,
      }}
      aria-hidden="true"
    >
      {name}
    </span>
  );
}