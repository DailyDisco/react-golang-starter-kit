import { useMemo } from "react";

import { Check, X } from "lucide-react";

interface PasswordStrengthMeterProps {
  password: string;
}

interface PasswordRequirement {
  label: string;
  test: (password: string) => boolean;
}

const requirements: PasswordRequirement[] = [
  { label: "At least 8 characters", test: (p) => p.length >= 8 },
  { label: "Contains uppercase letter", test: (p) => /[A-Z]/.test(p) },
  { label: "Contains lowercase letter", test: (p) => /[a-z]/.test(p) },
  { label: "Contains number", test: (p) => /\d/.test(p) },
  { label: "Contains special character", test: (p) => /[!@#$%^&*(),.?":{}|<>]/.test(p) },
];

function getStrengthLevel(score: number): {
  label: string;
  color: string;
  bgColor: string;
} {
  if (score === 0) return { label: "", color: "bg-muted", bgColor: "bg-muted" };
  if (score <= 2) return { label: "Weak", color: "bg-red-500", bgColor: "bg-red-100" };
  if (score <= 3) return { label: "Fair", color: "bg-yellow-500", bgColor: "bg-yellow-100" };
  if (score <= 4) return { label: "Good", color: "bg-blue-500", bgColor: "bg-blue-100" };
  return { label: "Strong", color: "bg-green-500", bgColor: "bg-green-100" };
}

/**
 * PasswordStrengthMeter provides real-time visual feedback on password strength.
 * It displays a progress bar and checklist of security requirements.
 */
export function PasswordStrengthMeter({ password }: PasswordStrengthMeterProps) {
  const { score, passedRequirements } = useMemo(() => {
    const passed = requirements.map((req) => req.test(password));
    return {
      score: passed.filter(Boolean).length,
      passedRequirements: passed,
    };
  }, [password]);

  const strength = getStrengthLevel(score);

  // Don't show anything if password is empty
  if (!password) return null;

  return (
    <div className="mt-2 space-y-2">
      {/* Strength bar */}
      <div className="flex items-center gap-2">
        <div className="bg-muted h-2 flex-1 overflow-hidden rounded-full">
          <div
            className={`h-full transition-all duration-300 ${strength.color}`}
            style={{ width: `${(score / requirements.length) * 100}%` }}
          />
        </div>
        {strength.label && (
          <span
            className={`text-xs font-medium ${
              score <= 2
                ? "text-red-600"
                : score <= 3
                  ? "text-yellow-600"
                  : score <= 4
                    ? "text-blue-600"
                    : "text-green-600"
            }`}
          >
            {strength.label}
          </span>
        )}
      </div>

      {/* Requirements checklist */}
      <ul className="space-y-1 text-xs">
        {requirements.map((req, index) => (
          <li
            key={req.label}
            className="flex items-center gap-1.5"
          >
            {passedRequirements[index] ? (
              <Check
                className="h-3 w-3 text-green-600"
                aria-hidden="true"
              />
            ) : (
              <X
                className="text-muted-foreground h-3 w-3"
                aria-hidden="true"
              />
            )}
            <span className={passedRequirements[index] ? "text-green-600" : "text-muted-foreground"}>{req.label}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
