import { useMemo } from "react";

import { Check, X } from "lucide-react";
import { useTranslation } from "react-i18next";

interface PasswordStrengthMeterProps {
  password: string;
}

interface PasswordRequirement {
  key: string;
  test: (password: string) => boolean;
}

const requirementTests: PasswordRequirement[] = [
  { key: "length", test: (p) => p.length >= 8 },
  { key: "uppercase", test: (p) => /[A-Z]/.test(p) },
  { key: "lowercase", test: (p) => /[a-z]/.test(p) },
  { key: "number", test: (p) => /\d/.test(p) },
  { key: "special", test: (p) => /[!@#$%^&*(),.?":{}|<>]/.test(p) },
];

type StrengthKey = "weak" | "fair" | "good" | "strong";

function getStrengthLevel(score: number): {
  key: StrengthKey | "";
  color: string;
  bgColor: string;
  textColor: string;
} {
  if (score === 0) return { key: "", color: "bg-muted", bgColor: "bg-muted", textColor: "text-muted-foreground" };
  if (score <= 2)
    return { key: "weak", color: "bg-destructive", bgColor: "bg-destructive/10", textColor: "text-destructive" };
  if (score <= 3) return { key: "fair", color: "bg-warning", bgColor: "bg-warning/10", textColor: "text-warning" };
  if (score <= 4) return { key: "good", color: "bg-info", bgColor: "bg-info/10", textColor: "text-info" };
  return { key: "strong", color: "bg-success", bgColor: "bg-success/10", textColor: "text-success" };
}

/**
 * PasswordStrengthMeter provides real-time visual feedback on password strength.
 * It displays a progress bar and checklist of security requirements.
 */
export function PasswordStrengthMeter({ password }: PasswordStrengthMeterProps) {
  const { t } = useTranslation("auth");

  const { score, passedRequirements } = useMemo(() => {
    const passed = requirementTests.map((req) => req.test(password));
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
            style={{ width: `${(score / requirementTests.length) * 100}%` }}
          />
        </div>
        {strength.key && (
          <span className={`text-xs font-medium ${strength.textColor}`}>{t(`passwordStrength.${strength.key}`)}</span>
        )}
      </div>

      {/* Requirements checklist */}
      <ul className="space-y-1 text-xs">
        {requirementTests.map((req, index) => (
          <li
            key={req.key}
            className="flex items-center gap-1.5"
          >
            {passedRequirements[index] ? (
              <Check
                className="text-success h-3 w-3"
                aria-hidden="true"
              />
            ) : (
              <X
                className="text-muted-foreground h-3 w-3"
                aria-hidden="true"
              />
            )}
            <span className={passedRequirements[index] ? "text-success" : "text-muted-foreground"}>
              {t(`passwordStrength.requirements.${req.key}` as `passwordStrength.requirements.length`)}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
