/**
 * Production-safe logging utility
 *
 * Provides structured logging that:
 * - Logs to console in development
 * - Suppresses logs in production (can be extended to send to logging service)
 * - Supports different log levels (debug, info, warn, error)
 */

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogContext {
  [key: string]: unknown;
}

// Check if we're in production mode
const isProduction = (): boolean => {
  if (typeof window === "undefined") {
    return process.env.NODE_ENV === "production";
  }
  return import.meta.env.PROD;
};

// Check if we're in development mode
const isDevelopment = (): boolean => {
  if (typeof window === "undefined") {
    return process.env.NODE_ENV === "development";
  }
  return import.meta.env.DEV;
};

// Determine if logging should be enabled
const shouldLog = (level: LogLevel): boolean => {
  // Always log errors
  if (level === "error") return true;

  // In production, only log errors and warnings
  if (isProduction()) {
    return level === "warn";
  }

  // In development, log everything
  return true;
};

// Format log message with timestamp and context
const formatMessage = (level: LogLevel, message: string, context?: LogContext): string => {
  const timestamp = new Date().toISOString();
  const contextStr = context ? ` ${JSON.stringify(context)}` : "";
  return `[${timestamp}] [${level.toUpperCase()}] ${message}${contextStr}`;
};

/**
 * Logger object with methods for each log level
 */
export const logger = {
  /**
   * Debug level logging - only shown in development
   */
  debug: (message: string, context?: LogContext): void => {
    if (shouldLog("debug")) {
      console.debug(formatMessage("debug", message, context));
    }
  },

  /**
   * Info level logging - only shown in development
   */
  info: (message: string, context?: LogContext): void => {
    if (shouldLog("info")) {
      console.info(formatMessage("info", message, context));
    }
  },

  /**
   * Warning level logging - shown in development and production
   */
  warn: (message: string, context?: LogContext): void => {
    if (shouldLog("warn")) {
      console.warn(formatMessage("warn", message, context));
    }
  },

  /**
   * Error level logging - always shown
   * In production, this could be extended to send to error tracking service
   */
  error: (message: string, error?: unknown, context?: LogContext): void => {
    if (shouldLog("error")) {
      const errorDetails = error instanceof Error ? { name: error.name, message: error.message } : { error };
      console.error(formatMessage("error", message, { ...errorDetails, ...context }));

      // In production, you could send to an error tracking service here
      // Example: Sentry.captureException(error, { extra: context });
    }
  },

  /**
   * Check if in development mode (useful for conditional logging)
   */
  isDev: isDevelopment,

  /**
   * Check if in production mode
   */
  isProd: isProduction,
};

export default logger;
