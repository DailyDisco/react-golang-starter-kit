import { useState } from "react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Textarea } from "@/components/ui/textarea";
import { queryKeys } from "@/lib/query-keys";
import type { User as UserType } from "@/services/types";
import { UserService } from "@/services/users/userService";
import { useAuthStore } from "@/stores/auth-store";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import {
  ArrowRight,
  Building2,
  Check,
  CheckCircle,
  ChevronRight,
  Rocket,
  Settings,
  Sparkles,
  User,
  Users,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/onboarding")({
  component: OnboardingPage,
});

type OnboardingStep = "welcome" | "profile" | "features" | "complete";

const STEPS: OnboardingStep[] = ["welcome", "profile", "features", "complete"];

function OnboardingPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user, login } = useAuthStore();
  const [currentStep, setCurrentStep] = useState<OnboardingStep>("welcome");
  const [profileData, setProfileData] = useState({
    name: user?.name || "",
    bio: "",
  });

  const currentStepIndex = STEPS.indexOf(currentStep);
  const progress = ((currentStepIndex + 1) / STEPS.length) * 100;

  const updateProfileMutation = useMutation<UserType, Error, { name?: string; bio?: string }>({
    mutationFn: (data) => UserService.updateCurrentUser(data),
    onSuccess: (updatedUser) => {
      login(updatedUser);
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
      toast.success(t("onboarding.profile.saved", "Profile updated successfully"));
      setCurrentStep("features");
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("onboarding.profile.error", "Failed to update profile"));
    },
  });

  const handleProfileSubmit = () => {
    if (profileData.name.trim()) {
      updateProfileMutation.mutate(profileData);
    } else {
      // Skip profile update if name hasn't changed
      setCurrentStep("features");
    }
  };

  const handleSkip = () => {
    const nextStepIndex = currentStepIndex + 1;
    if (nextStepIndex < STEPS.length) {
      setCurrentStep(STEPS[nextStepIndex]);
    }
  };

  const handleComplete = () => {
    navigate({ to: "/dashboard" });
  };

  return (
    <div className="flex min-h-[calc(100vh-4rem)] flex-col items-center justify-center px-4 py-8">
      <div className="w-full max-w-2xl space-y-6">
        {/* Progress */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">
              {t("onboarding.step", "Step")} {currentStepIndex + 1} {t("onboarding.of", "of")} {STEPS.length}
            </span>
            <span className="font-medium">{Math.round(progress)}%</span>
          </div>
          <Progress
            value={progress}
            className="h-2"
          />
        </div>

        {/* Step Content */}
        {currentStep === "welcome" && (
          <WelcomeStep
            userName={user?.name || "there"}
            onContinue={() => setCurrentStep("profile")}
          />
        )}

        {currentStep === "profile" && (
          <ProfileStep
            profileData={profileData}
            setProfileData={setProfileData}
            onSubmit={handleProfileSubmit}
            onSkip={handleSkip}
            isLoading={updateProfileMutation.isPending}
          />
        )}

        {currentStep === "features" && (
          <FeaturesStep
            onContinue={() => setCurrentStep("complete")}
            onSkip={handleSkip}
          />
        )}

        {currentStep === "complete" && <CompleteStep onComplete={handleComplete} />}
      </div>
    </div>
  );
}

function WelcomeStep({ userName, onContinue }: { userName: string; onContinue: () => void }) {
  const { t } = useTranslation();

  return (
    <Card className="border-2">
      <CardHeader className="text-center">
        <div className="bg-primary/10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
          <Sparkles className="text-primary h-8 w-8" />
        </div>
        <CardTitle className="text-2xl">
          {t("onboarding.welcome.title", "Welcome, {{name}}!", { name: userName })}
        </CardTitle>
        <CardDescription className="text-base">
          {t(
            "onboarding.welcome.description",
            "We're excited to have you here. Let's get you set up in just a few steps."
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 sm:grid-cols-3">
          <div className="flex flex-col items-center rounded-lg border p-4 text-center">
            <User className="text-muted-foreground mb-2 h-6 w-6" />
            <p className="text-sm font-medium">{t("onboarding.welcome.step1", "Complete your profile")}</p>
          </div>
          <div className="flex flex-col items-center rounded-lg border p-4 text-center">
            <Rocket className="text-muted-foreground mb-2 h-6 w-6" />
            <p className="text-sm font-medium">{t("onboarding.welcome.step2", "Explore key features")}</p>
          </div>
          <div className="flex flex-col items-center rounded-lg border p-4 text-center">
            <CheckCircle className="text-muted-foreground mb-2 h-6 w-6" />
            <p className="text-sm font-medium">{t("onboarding.welcome.step3", "Start using the app")}</p>
          </div>
        </div>

        <Button
          onClick={onContinue}
          className="w-full"
          size="lg"
        >
          {t("onboarding.welcome.getStarted", "Get Started")}
          <ArrowRight className="ml-2 h-4 w-4" />
        </Button>
      </CardContent>
    </Card>
  );
}

function ProfileStep({
  profileData,
  setProfileData,
  onSubmit,
  onSkip,
  isLoading,
}: {
  profileData: { name: string; bio: string };
  setProfileData: (data: { name: string; bio: string }) => void;
  onSubmit: () => void;
  onSkip: () => void;
  isLoading: boolean;
}) {
  const { t } = useTranslation();
  const { user } = useAuthStore();

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-4">
          <Avatar className="h-16 w-16">
            <AvatarFallback className="text-lg">{profileData.name?.charAt(0)?.toUpperCase() || "?"}</AvatarFallback>
          </Avatar>
          <div>
            <CardTitle>{t("onboarding.profile.title", "Complete Your Profile")}</CardTitle>
            <CardDescription>
              {t("onboarding.profile.description", "Tell us a bit about yourself so others can recognize you.")}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-2">
          <Label htmlFor="name">{t("onboarding.profile.name", "Display Name")}</Label>
          <Input
            id="name"
            value={profileData.name}
            onChange={(e) => setProfileData({ ...profileData, name: e.target.value })}
            placeholder={t("onboarding.profile.namePlaceholder", "How should we call you?")}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="bio">
            {t("onboarding.profile.bio", "Bio")} <span className="text-muted-foreground">(optional)</span>
          </Label>
          <Textarea
            id="bio"
            value={profileData.bio}
            onChange={(e) => setProfileData({ ...profileData, bio: e.target.value })}
            placeholder={t("onboarding.profile.bioPlaceholder", "Tell us a little about yourself...")}
            rows={3}
          />
        </div>

        {user?.email && (
          <div className="space-y-2">
            <Label className="text-muted-foreground">{t("onboarding.profile.email", "Email")}</Label>
            <p className="text-sm">{user.email}</p>
          </div>
        )}

        <div className="flex gap-3">
          <Button
            variant="outline"
            onClick={onSkip}
            className="flex-1"
            disabled={isLoading}
          >
            {t("common.skip", "Skip for now")}
          </Button>
          <Button
            onClick={onSubmit}
            className="flex-1"
            disabled={isLoading}
          >
            {isLoading ? t("common.saving", "Saving...") : t("common.continue", "Continue")}
            <ChevronRight className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function FeaturesStep({ onContinue, onSkip }: { onContinue: () => void; onSkip: () => void }) {
  const { t } = useTranslation();

  const features = [
    {
      icon: <Building2 className="h-5 w-5" />,
      title: t("onboarding.features.organizations.title", "Organizations"),
      description: t(
        "onboarding.features.organizations.description",
        "Create or join organizations to collaborate with your team."
      ),
      link: "/settings",
      badge: "Core",
    },
    {
      icon: <Users className="h-5 w-5" />,
      title: t("onboarding.features.team.title", "Team Management"),
      description: t(
        "onboarding.features.team.description",
        "Invite team members, assign roles, and manage permissions."
      ),
      badge: "Collaboration",
    },
    {
      icon: <Settings className="h-5 w-5" />,
      title: t("onboarding.features.settings.title", "Personalization"),
      description: t(
        "onboarding.features.settings.description",
        "Customize your experience with themes, notifications, and preferences."
      ),
      link: "/settings/preferences",
      badge: "Settings",
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("onboarding.features.title", "Key Features")}</CardTitle>
        <CardDescription>
          {t("onboarding.features.description", "Here's what you can do with the platform.")}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-4">
          {features.map((feature, index) => (
            <div
              key={index}
              className="hover:bg-accent/50 flex items-start gap-4 rounded-lg border p-4 transition-colors"
            >
              <div className="bg-primary/10 text-primary flex h-10 w-10 shrink-0 items-center justify-center rounded-lg">
                {feature.icon}
              </div>
              <div className="flex-1 space-y-1">
                <div className="flex items-center gap-2">
                  <p className="font-medium">{feature.title}</p>
                  <Badge
                    variant="secondary"
                    className="text-xs"
                  >
                    {feature.badge}
                  </Badge>
                </div>
                <p className="text-muted-foreground text-sm">{feature.description}</p>
              </div>
              {feature.link && (
                <Link
                  to={feature.link}
                  className="text-primary text-sm hover:underline"
                >
                  {t("common.explore", "Explore")}
                </Link>
              )}
            </div>
          ))}
        </div>

        <div className="flex gap-3">
          <Button
            variant="outline"
            onClick={onSkip}
            className="flex-1"
          >
            {t("common.skip", "Skip")}
          </Button>
          <Button
            onClick={onContinue}
            className="flex-1"
          >
            {t("common.gotIt", "Got it!")}
            <Check className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function CompleteStep({ onComplete }: { onComplete: () => void }) {
  const { t } = useTranslation();

  return (
    <Card className="border-2 border-green-500/30 bg-green-50/50 dark:bg-green-950/10">
      <CardHeader className="text-center">
        <div className="bg-success/10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
          <CheckCircle className="text-success h-8 w-8" />
        </div>
        <CardTitle className="text-2xl">{t("onboarding.complete.title", "You're All Set!")}</CardTitle>
        <CardDescription className="text-base">
          {t(
            "onboarding.complete.description",
            "Your account is ready. Start exploring the platform and make the most of it."
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 sm:grid-cols-2">
          <Link to="/dashboard">
            <Card className="hover:bg-accent/50 cursor-pointer transition-colors">
              <CardContent className="flex items-center gap-3 p-4">
                <Rocket className="text-primary h-5 w-5" />
                <div>
                  <p className="font-medium">{t("onboarding.complete.dashboard", "Go to Dashboard")}</p>
                  <p className="text-muted-foreground text-xs">
                    {t("onboarding.complete.dashboardDesc", "See your overview")}
                  </p>
                </div>
              </CardContent>
            </Card>
          </Link>
          <Link to="/settings/profile">
            <Card className="hover:bg-accent/50 cursor-pointer transition-colors">
              <CardContent className="flex items-center gap-3 p-4">
                <Settings className="text-primary h-5 w-5" />
                <div>
                  <p className="font-medium">{t("onboarding.complete.settings", "Complete Profile")}</p>
                  <p className="text-muted-foreground text-xs">
                    {t("onboarding.complete.settingsDesc", "Add more details")}
                  </p>
                </div>
              </CardContent>
            </Card>
          </Link>
        </div>

        <Button
          onClick={onComplete}
          className="w-full"
          size="lg"
        >
          {t("onboarding.complete.button", "Go to Dashboard")}
          <ArrowRight className="ml-2 h-4 w-4" />
        </Button>
      </CardContent>
    </Card>
  );
}
