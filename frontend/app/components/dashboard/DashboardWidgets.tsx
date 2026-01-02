import { memo, useCallback, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronUp, GripVertical, Maximize2, Minimize2, X } from "lucide-react";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface Widget {
  id: string;
  title: string;
  component: React.ComponentType<{ isCollapsed?: boolean }>;
  defaultCollapsed?: boolean;
  removable?: boolean;
  /** Size: 1 = 1/3 width, 2 = 2/3 width, 3 = full width */
  size?: 1 | 2 | 3;
}

interface WidgetState {
  /** Collapsed state for each widget */
  collapsed: Record<string, boolean>;
  /** Order of widgets by ID */
  order: string[];
  /** Hidden widgets */
  hidden: string[];
  /** Toggle collapsed state */
  toggleCollapsed: (id: string) => void;
  /** Set collapsed state */
  setCollapsed: (id: string, collapsed: boolean) => void;
  /** Reorder widgets */
  reorder: (fromIndex: number, toIndex: number) => void;
  /** Hide a widget */
  hideWidget: (id: string) => void;
  /** Show a widget */
  showWidget: (id: string) => void;
  /** Reset to defaults */
  reset: (widgets: Widget[]) => void;
}

export const useWidgetStore = create<WidgetState>()(
  persist(
    (set) => ({
      collapsed: {},
      order: [],
      hidden: [],

      toggleCollapsed: (id) =>
        set((state) => ({
          collapsed: { ...state.collapsed, [id]: !state.collapsed[id] },
        })),

      setCollapsed: (id, collapsed) =>
        set((state) => ({
          collapsed: { ...state.collapsed, [id]: collapsed },
        })),

      reorder: (fromIndex, toIndex) =>
        set((state) => {
          const newOrder = [...state.order];
          const [removed] = newOrder.splice(fromIndex, 1);
          newOrder.splice(toIndex, 0, removed);
          return { order: newOrder };
        }),

      hideWidget: (id) =>
        set((state) => ({
          hidden: [...new Set([...state.hidden, id])],
        })),

      showWidget: (id) =>
        set((state) => ({
          hidden: state.hidden.filter((h) => h !== id),
        })),

      reset: (widgets) =>
        set({
          collapsed: widgets.reduce(
            (acc, w) => ({ ...acc, [w.id]: w.defaultCollapsed || false }),
            {} as Record<string, boolean>
          ),
          order: widgets.map((w) => w.id),
          hidden: [],
        }),
    }),
    {
      name: "dashboard-widgets",
    }
  )
);

interface DashboardWidgetProps {
  widget: Widget;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
  onHide?: () => void;
  isDragging?: boolean;
  onDragStart?: () => void;
  onDragEnd?: () => void;
  onDragOver?: () => void;
}

const DashboardWidget = memo(function DashboardWidget({
  widget,
  isCollapsed,
  onToggleCollapse,
  onHide,
  isDragging,
  onDragStart,
  onDragEnd,
  onDragOver,
}: DashboardWidgetProps) {
  const WidgetComponent = widget.component;

  return (
    <Card
      className={cn(
        "transition-all",
        isDragging && "ring-primary opacity-50 ring-2",
        widget.size === 1 && "lg:col-span-1",
        widget.size === 2 && "lg:col-span-2",
        widget.size === 3 && "lg:col-span-3"
      )}
      draggable
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
      onDragOver={(e) => {
        e.preventDefault();
        onDragOver?.();
      }}
    >
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <GripVertical className="text-muted-foreground h-4 w-4 cursor-grab" />
          <CardTitle className="text-base font-medium">{widget.title}</CardTitle>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={onToggleCollapse}
          >
            {isCollapsed ? <Maximize2 className="h-4 w-4" /> : <Minimize2 className="h-4 w-4" />}
          </Button>
          {widget.removable !== false && onHide && (
            <Button
              variant="ghost"
              size="icon"
              className="text-muted-foreground hover:text-destructive h-7 w-7"
              onClick={onHide}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </CardHeader>
      {!isCollapsed && (
        <CardContent>
          <WidgetComponent isCollapsed={isCollapsed} />
        </CardContent>
      )}
    </Card>
  );
});

interface DashboardWidgetsProps {
  /** Available widgets */
  widgets: Widget[];
  /** CSS class for the grid container */
  className?: string;
}

/**
 * Dashboard with customizable, reorderable widgets
 *
 * @example
 * <DashboardWidgets
 *   widgets={[
 *     { id: "activity", title: "Activity", component: ActivityWidget },
 *     { id: "stats", title: "Stats", component: StatsWidget, size: 2 },
 *   ]}
 * />
 */
export function DashboardWidgets({ widgets, className }: DashboardWidgetsProps) {
  const { collapsed, order, hidden, toggleCollapsed, hideWidget, showWidget, reorder, reset } = useWidgetStore();
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

  // Initialize order if empty
  if (order.length === 0) {
    reset(widgets);
  }

  // Sort widgets by stored order
  const sortedWidgets = [...widgets].sort((a, b) => {
    const aIndex = order.indexOf(a.id);
    const bIndex = order.indexOf(b.id);
    if (aIndex === -1) return 1;
    if (bIndex === -1) return -1;
    return aIndex - bIndex;
  });

  // Filter out hidden widgets
  const visibleWidgets = sortedWidgets.filter((w) => !hidden.includes(w.id));
  const hiddenWidgets = widgets.filter((w) => hidden.includes(w.id));

  const handleDragStart = useCallback((index: number) => {
    setDraggedIndex(index);
  }, []);

  const handleDragEnd = useCallback(() => {
    setDraggedIndex(null);
  }, []);

  const handleDragOver = useCallback(
    (targetIndex: number) => {
      if (draggedIndex === null || draggedIndex === targetIndex) return;

      const draggedWidgetId = visibleWidgets[draggedIndex]?.id;
      const targetWidgetId = visibleWidgets[targetIndex]?.id;

      if (!draggedWidgetId || !targetWidgetId) return;

      const fromOrderIndex = order.indexOf(draggedWidgetId);
      const toOrderIndex = order.indexOf(targetWidgetId);

      if (fromOrderIndex !== -1 && toOrderIndex !== -1) {
        reorder(fromOrderIndex, toOrderIndex);
        setDraggedIndex(targetIndex);
      }
    },
    [draggedIndex, visibleWidgets, order, reorder]
  );

  return (
    <div className={cn("space-y-4", className)}>
      {/* Hidden widgets bar */}
      {hiddenWidgets.length > 0 && (
        <div className="bg-muted/50 flex flex-wrap items-center gap-2 rounded-lg p-2">
          <span className="text-muted-foreground text-sm">Hidden widgets:</span>
          {hiddenWidgets.map((widget) => (
            <Button
              key={widget.id}
              variant="outline"
              size="sm"
              className="h-7 gap-1"
              onClick={() => showWidget(widget.id)}
            >
              {widget.title}
              <ChevronUp className="h-3 w-3" />
            </Button>
          ))}
        </div>
      )}

      {/* Widget grid */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {visibleWidgets.map((widget, index) => (
          <DashboardWidget
            key={widget.id}
            widget={widget}
            isCollapsed={collapsed[widget.id] || false}
            onToggleCollapse={() => toggleCollapsed(widget.id)}
            onHide={widget.removable !== false ? () => hideWidget(widget.id) : undefined}
            isDragging={draggedIndex === index}
            onDragStart={() => handleDragStart(index)}
            onDragEnd={handleDragEnd}
            onDragOver={() => handleDragOver(index)}
          />
        ))}
      </div>

      {/* Collapse/Expand all */}
      <div className="flex justify-end gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            for (const w of visibleWidgets) useWidgetStore.getState().setCollapsed(w.id, true);
          }}
        >
          <ChevronUp className="mr-1 h-4 w-4" />
          Collapse All
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            for (const w of visibleWidgets) useWidgetStore.getState().setCollapsed(w.id, false);
          }}
        >
          <ChevronDown className="mr-1 h-4 w-4" />
          Expand All
        </Button>
      </div>
    </div>
  );
}

/**
 * Hook for widget customization
 */
export function useWidgets() {
  const store = useWidgetStore();
  return {
    collapsed: store.collapsed,
    order: store.order,
    hidden: store.hidden,
    toggleCollapsed: store.toggleCollapsed,
    hideWidget: store.hideWidget,
    showWidget: store.showWidget,
    reorder: store.reorder,
    reset: store.reset,
  };
}
