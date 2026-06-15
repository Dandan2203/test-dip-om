import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { ArrowUpRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { useDashboardEdit } from "@/components/dashboard/editContext";

interface WidgetFrameProps {
  title?: string;
  /** Куди веде заголовок-посилання (напр. "/app/goals"). У режимі редагування неактивне. */
  to?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
}

export function WidgetFrame({ title, to, action, children, className }: WidgetFrameProps) {
  const editMode = useDashboardEdit();
  const linked = to && !editMode;

  return (
    <div className={cn("flex h-full flex-col overflow-hidden rounded-2xl border border-border bg-card", className)}>
      {title && (
        <div className="flex shrink-0 items-center justify-between px-4 pt-3 pb-1">
          {linked ? (
            <Link
              to={to}
              className="group flex items-center gap-1 text-sm font-semibold text-muted-foreground transition-colors hover:text-primary"
            >
              {title}
              <ArrowUpRight
                size={13}
                className="opacity-0 transition-opacity group-hover:opacity-100"
              />
            </Link>
          ) : (
            <h3 className="text-sm font-semibold text-muted-foreground">{title}</h3>
          )}
          {action}
        </div>
      )}
      <div className="min-h-0 flex-1 overflow-auto px-4 pb-4 pt-1">{children}</div>
    </div>
  );
}
