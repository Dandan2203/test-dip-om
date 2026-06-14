import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface WidgetFrameProps {
  title?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
}

export function WidgetFrame({ title, action, children, className }: WidgetFrameProps) {
  return (
    <div className={cn("flex h-full flex-col overflow-hidden rounded-2xl border border-border bg-card", className)}>
      {title && (
        <div className="flex shrink-0 items-center justify-between px-4 pt-3 pb-1">
          <h3 className="text-sm font-semibold text-muted-foreground">{title}</h3>
          {action}
        </div>
      )}
      <div className="min-h-0 flex-1 overflow-auto px-4 pb-4 pt-1">{children}</div>
    </div>
  );
}
