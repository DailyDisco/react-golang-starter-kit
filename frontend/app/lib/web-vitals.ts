import type { Metric } from "web-vitals";

import { addBreadcrumb, isSentryConfigured } from "./sentry";

/**
 * Rating thresholds based on Google's Core Web Vitals recommendations
 */
const getRating = (name: string, value: number): "good" | "needs-improvement" | "poor" => {
  const thresholds: Record<string, [number, number]> = {
    CLS: [0.1, 0.25],
    FID: [100, 300],
    FCP: [1800, 3000],
    LCP: [2500, 4000],
    TTFB: [800, 1800],
    INP: [200, 500],
  };

  const [good, poor] = thresholds[name] || [0, 0];
  if (value <= good) return "good";
  if (value <= poor) return "needs-improvement";
  return "poor";
};

/**
 * Report Web Vital metric to Sentry as a breadcrumb
 */
function reportToSentry(metric: Metric): void {
  const rating = getRating(metric.name, metric.value);

  addBreadcrumb({
    message: `Web Vital: ${metric.name}`,
    category: "web-vitals",
    data: {
      name: metric.name,
      value: Math.round(metric.name === "CLS" ? metric.value * 1000 : metric.value),
      rating,
      delta: Math.round(metric.delta),
      id: metric.id,
      navigationType: metric.navigationType,
    },
  });
}

/**
 * Initialize Web Vitals tracking
 * Reports metrics to Sentry as breadcrumbs for debugging
 */
export function initWebVitals(): void {
  if (typeof window === "undefined") return;

  // Only track if Sentry is configured
  if (!isSentryConfigured()) return;

  // Dynamically import web-vitals to avoid bundling if not used
  import("web-vitals").then(({ onCLS, onFCP, onLCP, onTTFB, onINP }) => {
    onCLS(reportToSentry);
    onFCP(reportToSentry);
    onLCP(reportToSentry);
    onTTFB(reportToSentry);
    onINP(reportToSentry);
  });
}
