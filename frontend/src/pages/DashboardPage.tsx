import { useState } from "react";
import { Pencil, Check } from "lucide-react";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";

export function DashboardPage() {
  const [dashboard, setDashboard] = useState("overview");
  const [editMode, setEditMode] = useState(false);

  return (
    <div className="mx-auto max-w-7xl space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Tabs value={dashboard} onValueChange={setDashboard}>
          <TabsList>
            <TabsTrigger value="overview">Огляд</TabsTrigger>
            <TabsTrigger value="goals">Цілі</TabsTrigger>
          </TabsList>
        </Tabs>
        <Button variant={editMode ? "primary" : "outline"} size="sm" onClick={() => setEditMode((v) => !v)}>
          {editMode ? <><Check size={15} /> Готово</> : <><Pencil size={15} /> Редагувати</>}
        </Button>
      </div>

      <DashboardGrid key={dashboard} name={dashboard} editMode={editMode} />
    </div>
  );
}
