import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type Variant = "default" | "primary" | "positive" | "negative" | "muted";

const variants: Record<Variant, string> = {
  default: "bg-muted text-muted-foreground",
  primary: "bg-primary-soft text-primary",
  positive: "bg-positive/10 text-positive",
  negative: "bg-negative/10 text-negative",
  muted: "bg-muted text-muted-foreground",
};

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: Variant;
}

export function Badge({ className, variant = "default", ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
        variants[variant],
        className,
      )}
      {...props}
    />
  );
}
