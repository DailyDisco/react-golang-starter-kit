import { useState } from "react";

import { Link } from "@tanstack/react-router";
import { AnimatePresence, motion } from "framer-motion";
import {
  Activity,
  Bell,
  Brain,
  Building2,
  ChevronLeft,
  ChevronRight,
  Code,
  CreditCard,
  FileUp,
  Flag,
  Globe,
  Image,
  Key,
  LayoutDashboard,
  Lock,
  Mail,
  MessageSquare,
  Radio,
  ScrollText,
  Send,
  Settings,
  Shield,
  Smartphone,
  Sparkles,
  Users,
  Zap,
} from "lucide-react";
import { toast } from "sonner";

import { useFileDelete, useFileUpload } from "../../hooks/mutations/use-file-mutations";
import { useCreateUser, useDeleteUser } from "../../hooks/mutations/use-user-mutations";
import { useFileDownload, useFiles, useStorageStatus } from "../../hooks/queries/use-files";
import { useHealthCheck } from "../../hooks/queries/use-health";
import { useUsers } from "../../hooks/queries/use-users";
import { API_BASE_URL, type FileResponse, type StorageStatus } from "../../services";
import { useAuthStore } from "../../stores/auth-store";
import { useFileStore } from "../../stores/file-store";
import { useUserStore } from "../../stores/user-store";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../ui/alert-dialog";

interface Step {
  id: string;
  title: string;
  icon: React.ReactNode;
  description: string;
}

const steps: Step[] = [
  {
    id: "auth",
    title: "Authentication",
    icon: <Key className="h-5 w-5" />,
    description: "Secure login with OAuth, 2FA, and session management",
  },
  {
    id: "organizations",
    title: "Organizations",
    icon: <Building2 className="h-5 w-5" />,
    description: "Multi-tenant teams with roles and invitations",
  },
  {
    id: "billing",
    title: "Billing",
    icon: <CreditCard className="h-5 w-5" />,
    description: "Stripe integration for subscriptions and payments",
  },
  {
    id: "admin",
    title: "Admin Panel",
    icon: <LayoutDashboard className="h-5 w-5" />,
    description: "Full admin dashboard with user management",
  },
  {
    id: "realtime",
    title: "Real-Time",
    icon: <Radio className="h-5 w-5" />,
    description: "WebSocket notifications and live data sync",
  },
  {
    id: "i18n",
    title: "i18n",
    icon: <Globe className="h-5 w-5" />,
    description: "Multi-language support with English and Spanish",
  },
  {
    id: "files",
    title: "File Storage",
    icon: <FileUp className="h-5 w-5" />,
    description: "Upload files to S3 or database storage",
  },
  {
    id: "health",
    title: "System Health",
    icon: <Activity className="h-5 w-5" />,
    description: "Real-time health monitoring with Prometheus metrics",
  },
  {
    id: "settings",
    title: "User Settings",
    icon: <Settings className="h-5 w-5" />,
    description: "Preferences, security, and account management",
  },
  {
    id: "ai",
    title: "AI Integration",
    icon: <Sparkles className="h-5 w-5" />,
    description: "Gemini-powered chat, vision, and embeddings",
  },
  {
    id: "developer",
    title: "Developer Experience",
    icon: <Code className="h-5 w-5" />,
    description: "API docs, hot reload, and modern tooling",
  },
];

export function Demo() {
  const [currentStep, setCurrentStep] = useState(0);

  // Server state
  const { data: _users, isLoading: _usersLoading } = useUsers();
  const { data: healthStatus, isLoading: healthLoading } = useHealthCheck();
  const { data: files, isLoading: filesLoading } = useFiles();
  const { data: storageStatus } = useStorageStatus();
  const { downloadFile } = useFileDownload();

  // Mutations
  const _createUserMutation = useCreateUser();
  const deleteUserMutation = useDeleteUser();
  const fileUploadMutation = useFileUpload();
  const fileDeleteMutation = useFileDelete();

  // Client state
  const { formData: newUser, setFormData: _setNewUser } = useUserStore();
  const { selectedFile, isDragOver, setSelectedFile, setIsDragOver, resetFileSelection } = useFileStore();
  const { isAuthenticated } = useAuthStore();

  // Local state
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState<{ id: number; name: string } | null>(null);

  // Password validation
  const passwordValidation = {
    length: newUser.password.length >= 8,
    uppercase: /[A-Z]/.test(newUser.password),
    lowercase: /[a-z]/.test(newUser.password),
  };
  const _isPasswordValid = passwordValidation.length && passwordValidation.uppercase && passwordValidation.lowercase;

  // Handlers
  const testHealthCheck = () => {
    if (healthStatus) {
      toast.success("Health check successful!", {
        description: `Overall: ${String(healthStatus.overall_status).toUpperCase()}`,
      });
    }
  };

  const handleDeleteUser = () => {
    if (!userToDelete) return;
    deleteUserMutation.mutate(userToDelete.id, {
      onSuccess: () => {
        setDeleteDialogOpen(false);
        setUserToDelete(null);
      },
    });
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) setSelectedFile(file);
  };

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragOver(false);
    const file = event.dataTransfer.files[0];
    if (file) setSelectedFile(file);
  };

  const handleFileUpload = () => {
    if (!selectedFile) {
      toast.error("No file selected");
      return;
    }
    fileUploadMutation.mutate(selectedFile, { onSuccess: () => resetFileSelection() });
  };

  const nextStep = () => setCurrentStep((prev) => Math.min(prev + 1, steps.length - 1));
  const prevStep = () => setCurrentStep((prev) => Math.max(prev - 1, 0));

  return (
    <main className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Header */}
      <div className="border-b border-gray-200 bg-white px-4 py-8 dark:border-gray-700 dark:bg-gray-800">
        <div className="mx-auto max-w-5xl text-center">
          <h1 className="mb-2 text-3xl font-bold text-gray-900 dark:text-white">Interactive Feature Tour</h1>
          <p className="text-gray-600 dark:text-gray-300">Explore all the features built into this starter kit</p>
        </div>
      </div>

      {/* Step Navigation */}
      <div className="border-b border-gray-200 bg-white px-4 py-4 dark:border-gray-700 dark:bg-gray-800">
        <div className="mx-auto max-w-5xl">
          <div className="flex items-center justify-between gap-2 overflow-x-auto pb-2">
            {steps.map((step, index) => (
              <button
                key={step.id}
                onClick={() => setCurrentStep(index)}
                className={`flex min-w-[120px] flex-col items-center gap-1 rounded-lg px-3 py-2 text-sm transition-all ${
                  currentStep === index
                    ? "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300"
                    : "text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
                }`}
              >
                <div className={`rounded-full p-1.5 ${currentStep === index ? "bg-blue-200 dark:bg-blue-800" : ""}`}>
                  {step.icon}
                </div>
                <span className="font-medium">{step.title}</span>
              </button>
            ))}
          </div>
          {/* Progress bar */}
          <div className="mt-2 h-1 rounded-full bg-gray-200 dark:bg-gray-700">
            <div
              className="h-1 rounded-full bg-blue-600 transition-all duration-300"
              style={{ width: `${((currentStep + 1) / steps.length) * 100}%` }}
            />
          </div>
        </div>
      </div>

      {/* Step Content */}
      <div className="px-4 py-8">
        <div className="mx-auto max-w-5xl">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentStep}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              {/* Step Header */}
              <div className="mb-6">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-blue-100 p-2 text-blue-600 dark:bg-blue-900 dark:text-blue-400">
                    {steps[currentStep].icon}
                  </div>
                  <div>
                    <h2 className="text-2xl font-bold text-gray-900 dark:text-white">{steps[currentStep].title}</h2>
                    <p className="text-gray-600 dark:text-gray-400">{steps[currentStep].description}</p>
                  </div>
                </div>
              </div>

              {/* Step Content */}
              {currentStep === 0 && <AuthStep />}
              {currentStep === 1 && <OrganizationsStep />}
              {currentStep === 2 && <BillingStep />}
              {currentStep === 3 && <AdminStep />}
              {currentStep === 4 && <RealtimeStep />}
              {currentStep === 5 && <I18nStep />}
              {currentStep === 6 && (
                <FileStorageStep
                  files={files}
                  filesLoading={filesLoading}
                  storageStatus={storageStatus}
                  selectedFile={selectedFile}
                  isDragOver={isDragOver}
                  isAuthenticated={isAuthenticated}
                  fileUploadMutation={fileUploadMutation}
                  fileDeleteMutation={fileDeleteMutation}
                  handleFileSelect={handleFileSelect}
                  handleDragOver={handleDragOver}
                  handleDragLeave={handleDragLeave}
                  handleDrop={handleDrop}
                  handleFileUpload={handleFileUpload}
                  handleFileDownload={(id: number, name: string) => downloadFile(id, name)}
                  handleFileDelete={(id: number) => fileDeleteMutation.mutate(id)}
                  resetFileSelection={resetFileSelection}
                />
              )}
              {currentStep === 7 && (
                <HealthStep
                  healthStatus={healthStatus}
                  healthLoading={healthLoading}
                  testHealthCheck={testHealthCheck}
                />
              )}
              {currentStep === 8 && <SettingsStep />}
              {currentStep === 9 && <AIStep isAuthenticated={isAuthenticated} />}
              {currentStep === 10 && <DeveloperStep />}
            </motion.div>
          </AnimatePresence>

          {/* Navigation Buttons */}
          <div className="mt-8 flex items-center justify-between border-t border-gray-200 pt-6 dark:border-gray-700">
            <button
              onClick={prevStep}
              disabled={currentStep === 0}
              className="flex items-center gap-2 rounded-lg px-4 py-2 font-medium text-gray-600 transition-colors hover:bg-gray-100 disabled:opacity-50 disabled:hover:bg-transparent dark:text-gray-300 dark:hover:bg-gray-800"
            >
              <ChevronLeft className="h-5 w-5" />
              Previous
            </button>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              Step {currentStep + 1} of {steps.length}
            </span>
            <button
              onClick={nextStep}
              disabled={currentStep === steps.length - 1}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50 disabled:hover:bg-blue-600 dark:bg-blue-700 dark:hover:bg-blue-600"
            >
              Next
              <ChevronRight className="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>

      {/* Delete Dialog */}
      <AlertDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete user &quot;{userToDelete?.name}&quot;? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setDeleteDialogOpen(false)}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteUser}
              className="bg-red-600 hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </main>
  );
}

// Step Components

function AuthStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<Key className="h-5 w-5" />}
          title="Email & Password"
          description="Secure registration and login with JWT tokens, password hashing, and email verification."
        />
        <FeatureCard
          icon={<Users className="h-5 w-5" />}
          title="OAuth Integration"
          description="One-click login with Google and GitHub. Easy to add more providers."
        />
        <FeatureCard
          icon={<Smartphone className="h-5 w-5" />}
          title="Two-Factor Authentication"
          description="TOTP-based 2FA with QR code setup and backup codes for account recovery."
        />
        <FeatureCard
          icon={<Lock className="h-5 w-5" />}
          title="Session Management"
          description="View all active sessions, see device info, and revoke access from anywhere."
        />
      </div>
      <div className="rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-950/50">
        <p className="text-sm text-blue-700 dark:text-blue-300">
          <strong>Try it:</strong> Create an account using the Register page, then explore your security settings at{" "}
          <Link
            to="/settings/security"
            className="underline"
          >
            /settings/security
          </Link>
        </p>
      </div>
    </div>
  );
}

function OrganizationsStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<Building2 className="h-5 w-5" />}
          title="Multi-Tenant Architecture"
          description="Create isolated workspaces for different teams or customers with complete data separation."
        />
        <FeatureCard
          icon={<Users className="h-5 w-5" />}
          title="Team Management"
          description="Invite team members via email, manage roles, and track membership status."
        />
        <FeatureCard
          icon={<Shield className="h-5 w-5" />}
          title="Role-Based Permissions"
          description="Three-tier role system: Owner (full control), Admin (manage members), Member (access only)."
        />
        <FeatureCard
          icon={<Mail className="h-5 w-5" />}
          title="Invitation Workflow"
          description="Secure invitation tokens with expiration, accept/decline flow, and email notifications."
        />
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Organization Plans</h3>
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-lg border border-gray-200 p-4 dark:border-gray-600">
            <h4 className="font-medium text-gray-900 dark:text-white">Free</h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">Basic features for small teams</p>
          </div>
          <div className="rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-950/50">
            <h4 className="font-medium text-blue-700 dark:text-blue-300">Pro</h4>
            <p className="text-sm text-blue-600 dark:text-blue-400">Advanced analytics & more members</p>
          </div>
          <div className="rounded-lg border border-purple-200 bg-purple-50 p-4 dark:border-purple-800 dark:bg-purple-950/50">
            <h4 className="font-medium text-purple-700 dark:text-purple-300">Enterprise</h4>
            <p className="text-sm text-purple-600 dark:text-purple-400">Custom branding & API access</p>
          </div>
        </div>
      </div>
      <div className="rounded-lg border border-violet-200 bg-violet-50 p-4 dark:border-violet-800 dark:bg-violet-950/50">
        <p className="text-sm text-violet-700 dark:text-violet-300">
          <strong>Try it:</strong> Create an organization and manage your team at{" "}
          <Link
            to="/dashboard"
            className="underline"
          >
            /dashboard
          </Link>
        </p>
      </div>
    </div>
  );
}

function RealtimeStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<Radio className="h-5 w-5" />}
          title="WebSocket Connection"
          description="Persistent connection with automatic reconnection and exponential backoff."
        />
        <FeatureCard
          icon={<Bell className="h-5 w-5" />}
          title="Real-Time Notifications"
          description="Push notifications to users instantly for important events and updates."
        />
        <FeatureCard
          icon={<Activity className="h-5 w-5" />}
          title="Live Data Sync"
          description="Automatic cache invalidation when data changes, keeping UI always up-to-date."
        />
        <FeatureCard
          icon={<Users className="h-5 w-5" />}
          title="Broadcast Messages"
          description="Send system-wide announcements to all connected users simultaneously."
        />
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Message Types</h3>
        <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-300">
          <li className="flex items-center gap-2">
            <span className="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900 dark:text-blue-300">
              notification
            </span>
            <span>User-specific alerts and updates</span>
          </li>
          <li className="flex items-center gap-2">
            <span className="rounded bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900 dark:text-green-300">
              user_update
            </span>
            <span>Profile, preferences, or session changes</span>
          </li>
          <li className="flex items-center gap-2">
            <span className="rounded bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900 dark:text-purple-300">
              broadcast
            </span>
            <span>System-wide announcements</span>
          </li>
        </ul>
      </div>
      <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 dark:border-rose-800 dark:bg-rose-950/50">
        <p className="text-sm text-rose-700 dark:text-rose-300">
          <strong>How it works:</strong> Log in to automatically connect to WebSocket. Notifications appear in real-time
          without page refresh.
        </p>
      </div>
    </div>
  );
}

function I18nStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<Globe className="h-5 w-5" />}
          title="Multi-Language Support"
          description="Full internationalization with English and Spanish translations included."
        />
        <FeatureCard
          icon={<Settings className="h-5 w-5" />}
          title="Language Detection"
          description="Automatically detects browser language preference with localStorage persistence."
        />
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Translation Namespaces</h3>
        <div className="grid gap-3 md:grid-cols-2">
          <div className="flex items-center gap-2">
            <span className="rounded bg-gray-100 px-2 py-1 font-mono text-xs text-gray-700 dark:bg-gray-700 dark:text-gray-300">
              common
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">Buttons, labels, navigation</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="rounded bg-gray-100 px-2 py-1 font-mono text-xs text-gray-700 dark:bg-gray-700 dark:text-gray-300">
              auth
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">Login, register, passwords</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="rounded bg-gray-100 px-2 py-1 font-mono text-xs text-gray-700 dark:bg-gray-700 dark:text-gray-300">
              errors
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">Error messages and alerts</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="rounded bg-gray-100 px-2 py-1 font-mono text-xs text-gray-700 dark:bg-gray-700 dark:text-gray-300">
              validation
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">Form validation messages</span>
          </div>
        </div>
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Supported Languages</h3>
        <div className="flex gap-4">
          <div className="flex items-center gap-2 rounded-lg border border-gray-200 px-4 py-2 dark:border-gray-600">
            <span className="text-2xl">EN</span>
            <span className="text-gray-600 dark:text-gray-400">English</span>
          </div>
          <div className="flex items-center gap-2 rounded-lg border border-gray-200 px-4 py-2 dark:border-gray-600">
            <span className="text-2xl">ES</span>
            <span className="text-gray-600 dark:text-gray-400">Espanol</span>
          </div>
        </div>
      </div>
      <div className="rounded-lg border border-cyan-200 bg-cyan-50 p-4 dark:border-cyan-800 dark:bg-cyan-950/50">
        <p className="text-sm text-cyan-700 dark:text-cyan-300">
          <strong>Adding languages:</strong> Create new locale files in{" "}
          <code className="rounded bg-cyan-100 px-1 dark:bg-cyan-900">frontend/app/i18n/locales/</code> and update the
          i18n config.
        </p>
      </div>
    </div>
  );
}

function BillingStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<CreditCard className="h-5 w-5" />}
          title="Stripe Integration"
          description="Complete billing infrastructure with Stripe Checkout and Customer Portal."
        />
        <FeatureCard
          icon={<ScrollText className="h-5 w-5" />}
          title="Subscription Plans"
          description="Dynamic pricing fetched from Stripe. Support for multiple tiers and features."
        />
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">How It Works</h3>
        <ol className="space-y-3 text-sm text-gray-600 dark:text-gray-300">
          <li className="flex gap-3">
            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-600 dark:bg-blue-900 dark:text-blue-400">
              1
            </span>
            <span>
              Users browse pricing plans on the{" "}
              <Link
                to="/pricing"
                className="text-blue-600 underline dark:text-blue-400"
              >
                /pricing
              </Link>{" "}
              page
            </span>
          </li>
          <li className="flex gap-3">
            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-600 dark:bg-blue-900 dark:text-blue-400">
              2
            </span>
            <span>Clicking &quot;Subscribe&quot; creates a Stripe Checkout session</span>
          </li>
          <li className="flex gap-3">
            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-600 dark:bg-blue-900 dark:text-blue-400">
              3
            </span>
            <span>Stripe webhooks update subscription status in your database</span>
          </li>
          <li className="flex gap-3">
            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-600 dark:bg-blue-900 dark:text-blue-400">
              4
            </span>
            <span>
              Users manage billing via the Customer Portal at{" "}
              <Link
                to="/billing"
                className="text-blue-600 underline dark:text-blue-400"
              >
                /billing
              </Link>
            </span>
          </li>
        </ol>
      </div>
    </div>
  );
}

function AdminStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-3">
        <FeatureCard
          icon={<LayoutDashboard className="h-5 w-5" />}
          title="Dashboard"
          description="User stats, subscription metrics, and system overview."
        />
        <FeatureCard
          icon={<Users className="h-5 w-5" />}
          title="User Management"
          description="View, impersonate, activate/deactivate users."
        />
        <FeatureCard
          icon={<Flag className="h-5 w-5" />}
          title="Feature Flags"
          description="Control feature rollout with percentages and targeting."
        />
        <FeatureCard
          icon={<ScrollText className="h-5 w-5" />}
          title="Audit Logs"
          description="Track all actions with user, action, and timestamp."
        />
        <FeatureCard
          icon={<Bell className="h-5 w-5" />}
          title="Announcements"
          description="Create site-wide banners for all or specific users."
        />
        <FeatureCard
          icon={<Mail className="h-5 w-5" />}
          title="Email Templates"
          description="Customize transactional emails with preview."
        />
      </div>
      <div className="rounded-lg border border-purple-200 bg-purple-50 p-4 dark:border-purple-800 dark:bg-purple-950/50">
        <p className="text-sm text-purple-700 dark:text-purple-300">
          <strong>Admin Access:</strong> The admin panel is available at{" "}
          <Link
            to="/admin"
            className="underline"
          >
            /admin
          </Link>{" "}
          for users with admin or super_admin roles.
        </p>
      </div>
    </div>
  );
}

function FileStorageStep({
  files,
  filesLoading,
  storageStatus,
  selectedFile,
  isDragOver,
  isAuthenticated,
  fileUploadMutation,
  fileDeleteMutation: _fileDeleteMutation,
  handleFileSelect,
  handleDragOver,
  handleDragLeave,
  handleDrop,
  handleFileUpload,
  handleFileDownload,
  handleFileDelete,
  resetFileSelection: _resetFileSelection,
}: {
  files: FileResponse[] | undefined;
  filesLoading: boolean;
  storageStatus: StorageStatus | undefined;
  selectedFile: File | null;
  isDragOver: boolean;
  isAuthenticated: boolean;
  fileUploadMutation: { isPending: boolean };
  fileDeleteMutation: { mutate: (id: number) => void };
  handleFileSelect: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleDragOver: (event: React.DragEvent) => void;
  handleDragLeave: (event: React.DragEvent) => void;
  handleDrop: (event: React.DragEvent) => void;
  handleFileUpload: () => void;
  handleFileDownload: (id: number, name: string) => void;
  handleFileDelete: (id: number) => void;
  resetFileSelection: () => void;
}) {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <div className="space-y-4">
          <h3 className="font-semibold text-gray-900 dark:text-white">Upload a File</h3>
          <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`relative rounded-lg border-2 border-dashed p-6 text-center transition-colors ${
              isDragOver ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20" : "border-gray-300 dark:border-gray-600"
            }`}
          >
            <input
              type="file"
              onChange={handleFileSelect}
              className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
              disabled={fileUploadMutation.isPending}
            />
            <FileUp className="mx-auto h-10 w-10 text-gray-400" />
            <p className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
              {selectedFile ? selectedFile.name : "Drop files or click to browse"}
            </p>
            {selectedFile && <p className="text-xs text-gray-500">{(selectedFile.size / 1024).toFixed(2)} KB</p>}
            {storageStatus?.storage_type && (
              <p className="mt-2 text-xs text-gray-500">
                Storage: <span className="font-medium">{storageStatus.storage_type.toUpperCase()}</span>
              </p>
            )}
          </div>
          <button
            onClick={handleFileUpload}
            disabled={!selectedFile || fileUploadMutation.isPending || !isAuthenticated}
            className="w-full rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-gray-400"
          >
            {fileUploadMutation.isPending ? "Uploading..." : !isAuthenticated ? "Log in to upload" : "Upload File"}
          </button>
        </div>
        <div className="space-y-4">
          <h3 className="font-semibold text-gray-900 dark:text-white">Your Files</h3>
          <div className="max-h-[300px] space-y-2 overflow-y-auto rounded-lg border border-gray-200 p-3 dark:border-gray-700">
            {filesLoading ? (
              <p className="text-center text-sm text-gray-500">Loading...</p>
            ) : !files || files.length === 0 ? (
              <p className="text-center text-sm text-gray-500">No files uploaded</p>
            ) : (
              files.map((file: any) => (
                <div
                  key={file.id}
                  className="flex items-center justify-between rounded bg-gray-50 p-2 dark:bg-gray-800"
                >
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-gray-900 dark:text-white">{file.file_name}</p>
                    <p className="text-xs text-gray-500">{(file.file_size / 1024).toFixed(1)} KB</p>
                  </div>
                  <div className="flex gap-1">
                    <button
                      onClick={() => handleFileDownload(file.id, file.file_name)}
                      className="rounded bg-blue-600 px-2 py-1 text-xs text-white hover:bg-blue-700"
                    >
                      Download
                    </button>
                    <button
                      onClick={() => handleFileDelete(file.id)}
                      className="rounded bg-red-600 px-2 py-1 text-xs text-white hover:bg-red-700"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

interface HealthComponent {
  name: string;
  status: string;
  message?: string;
}

interface HealthStatus {
  overall_status: string;
  components?: HealthComponent[];
}

function HealthStep({
  healthStatus,
  healthLoading,
  testHealthCheck,
}: {
  healthStatus: HealthStatus | undefined;
  healthLoading: boolean;
  testHealthCheck: () => void;
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={testHealthCheck}
          disabled={healthLoading}
          className="rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700 disabled:bg-gray-400"
        >
          {healthLoading ? "Checking..." : "Run Health Check"}
        </button>
        {healthStatus && (
          <span
            className={`rounded-full px-3 py-1 text-sm font-medium ${
              healthStatus.overall_status === "healthy"
                ? "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300"
                : "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300"
            }`}
          >
            {healthStatus.overall_status.toUpperCase()}
          </span>
        )}
      </div>

      {healthStatus?.components && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {healthStatus.components.map((component: any, index: number) => (
            <div
              key={index}
              className={`rounded-lg border p-4 ${
                component.status === "healthy"
                  ? "border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950/50"
                  : "border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950/50"
              }`}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium text-gray-900 capitalize dark:text-white">{component.name}</span>
                <span className={component.status === "healthy" ? "text-green-600" : "text-red-600"}>
                  {component.status === "healthy" ? "Healthy" : "Unhealthy"}
                </span>
              </div>
              {component.message && (
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">{component.message}</p>
              )}
            </div>
          ))}
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
          <h3 className="mb-2 font-semibold text-gray-900 dark:text-white">Monitored Components</h3>
          <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
            <li>Database connectivity and query performance</li>
            <li>Redis cache status and memory usage</li>
            <li>File storage availability</li>
            <li>Background job queue status</li>
          </ul>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
          <h3 className="mb-2 font-semibold text-gray-900 dark:text-white">Prometheus Metrics</h3>
          <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
            <li>HTTP requests (count, duration, in-flight)</li>
            <li>WebSocket connections and messages</li>
            <li>Authentication attempts and user registrations</li>
            <li>Cache hits/misses, job processing times</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

function SettingsStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<Settings className="h-5 w-5" />}
          title="Preferences"
          description="Theme (light/dark/system), timezone, language, and date format."
        />
        <FeatureCard
          icon={<Shield className="h-5 w-5" />}
          title="Security"
          description="Change password, enable 2FA, manage sessions."
        />
        <FeatureCard
          icon={<Bell className="h-5 w-5" />}
          title="Notifications"
          description="Configure email notifications for updates, marketing, and security alerts."
        />
        <FeatureCard
          icon={<Users className="h-5 w-5" />}
          title="Connected Accounts"
          description="Link and unlink OAuth providers (Google, GitHub)."
        />
      </div>
      <div className="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-950/50">
        <p className="text-sm text-green-700 dark:text-green-300">
          <strong>Try it:</strong> Visit{" "}
          <Link
            to="/settings"
            className="underline"
          >
            /settings
          </Link>{" "}
          to explore all user preferences and account options.
        </p>
      </div>
    </div>
  );
}

function AIStep({ isAuthenticated }: { isAuthenticated: boolean }) {
  const [message, setMessage] = useState("");
  const [chatHistory, setChatHistory] = useState<Array<{ role: "user" | "assistant"; content: string }>>([]);
  const [isLoading, setIsLoading] = useState(false);

  const handleSendMessage = async () => {
    if (!message.trim() || !isAuthenticated) return;

    const userMessage = message.trim();
    setMessage("");
    setChatHistory((prev) => [...prev, { role: "user", content: userMessage }]);
    setIsLoading(true);

    try {
      // Dynamic import to avoid circular dependencies
      const { AIService } = await import("../../services/ai/aiService");
      const response = await AIService.chat(
        [...chatHistory, { role: "user" as const, content: userMessage }].map((m) => ({
          role: (m.role === "assistant" ? "model" : m.role) as "user" | "model" | "system" | "assistant",
          content: m.content,
        })),
        { maxTokens: 500 }
      );
      setChatHistory((prev) => [...prev, { role: "assistant", content: response.content }]);
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Failed to get response";
      if (errorMessage.includes("not available") || errorMessage.includes("503")) {
        toast.error("AI service not configured", {
          description: "Set GEMINI_API_KEY in your environment to enable AI features",
        });
      } else {
        toast.error("AI request failed", {
          description: errorMessage,
        });
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          icon={<MessageSquare className="h-5 w-5" />}
          title="Multi-Turn Chat"
          description="Have conversations with context. The AI remembers previous messages in the chat."
        />
        <FeatureCard
          icon={<Zap className="h-5 w-5" />}
          title="Streaming Responses"
          description="Real-time token streaming via Server-Sent Events for instant feedback."
        />
        <FeatureCard
          icon={<Image className="h-5 w-5" />}
          title="Image Analysis"
          description="Upload images and ask questions about them using multi-modal vision."
        />
        <FeatureCard
          icon={<Brain className="h-5 w-5" />}
          title="Embeddings"
          description="Generate vector embeddings for semantic search and similarity."
        />
      </div>

      {/* Interactive Chat Demo */}
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Try AI Chat</h3>

        {!isAuthenticated ? (
          <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-4 dark:border-yellow-800 dark:bg-yellow-950/50">
            <p className="text-sm text-yellow-700 dark:text-yellow-300">
              <strong>Login required:</strong> Please{" "}
              <Link
                to="/login"
                className="underline"
              >
                log in
              </Link>{" "}
              to try the AI chat demo.
            </p>
          </div>
        ) : (
          <>
            {/* Chat Messages */}
            <div className="mb-4 max-h-[300px] space-y-3 overflow-y-auto rounded-lg border border-gray-100 bg-gray-50 p-4 dark:border-gray-600 dark:bg-gray-900">
              {chatHistory.length === 0 ? (
                <p className="text-center text-sm text-gray-500 dark:text-gray-400">
                  Send a message to start chatting with Gemini AI
                </p>
              ) : (
                chatHistory.map((msg, idx) => (
                  <div
                    key={idx}
                    className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
                  >
                    <div
                      className={`max-w-[80%] rounded-lg px-4 py-2 ${
                        msg.role === "user"
                          ? "bg-blue-600 text-white"
                          : "bg-white text-gray-900 dark:bg-gray-800 dark:text-white"
                      }`}
                    >
                      <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                    </div>
                  </div>
                ))
              )}
              {isLoading && (
                <div className="flex justify-start">
                  <div className="rounded-lg bg-white px-4 py-2 dark:bg-gray-800">
                    <div className="flex items-center gap-2 text-sm text-gray-500">
                      <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-blue-600"></span>
                      Thinking...
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Input */}
            <div className="flex gap-2">
              <input
                type="text"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && handleSendMessage()}
                placeholder="Type a message..."
                disabled={isLoading}
                className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-900 dark:text-white"
              />
              <button
                onClick={handleSendMessage}
                disabled={isLoading || !message.trim()}
                className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-gray-400"
              >
                <Send className="h-4 w-4" />
              </button>
            </div>
          </>
        )}
      </div>

      <div className="rounded-lg border border-fuchsia-200 bg-fuchsia-50 p-4 dark:border-fuchsia-800 dark:bg-fuchsia-950/50">
        <p className="text-sm text-fuchsia-700 dark:text-fuchsia-300">
          <strong>Configuration:</strong> Set{" "}
          <code className="rounded bg-fuchsia-100 px-1 dark:bg-fuchsia-900">GEMINI_API_KEY</code> in your environment to
          enable AI features. Get your API key from{" "}
          <a
            href="https://aistudio.google.com/app/apikey"
            target="_blank"
            rel="noopener noreferrer"
            className="underline"
          >
            Google AI Studio
          </a>
          .
        </p>
      </div>
    </div>
  );
}

function DeveloperStep() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
          <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Quick Commands</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-300">Start all services:</span>
              <code className="rounded bg-gray-100 px-2 py-1 dark:bg-gray-700">docker-compose up</code>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-300">View logs:</span>
              <code className="rounded bg-gray-100 px-2 py-1 dark:bg-gray-700">docker-compose logs -f</code>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-300">Run migrations:</span>
              <code className="rounded bg-gray-100 px-2 py-1 dark:bg-gray-700">make migrate-up</code>
            </div>
          </div>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
          <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Useful Links</h3>
          <ul className="space-y-2 text-sm">
            <li>
              <a
                href={`${API_BASE_URL}/swagger/`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline dark:text-blue-400"
              >
                API Documentation (Swagger)
              </a>
            </li>
            <li>
              <span className="text-gray-600 dark:text-gray-300">Frontend: localhost:5173</span>
            </li>
            <li>
              <span className="text-gray-600 dark:text-gray-300">Backend: localhost:8080</span>
            </li>
            <li>
              <span className="text-gray-600 dark:text-gray-300">Prometheus: localhost:9090</span>
            </li>
            <li>
              <span className="text-gray-600 dark:text-gray-300">Grafana: localhost:3000</span>
            </li>
          </ul>
        </div>
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Tech Stack</h3>
        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <h4 className="mb-2 text-sm font-medium text-blue-600 dark:text-blue-400">Frontend</h4>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>React 19 + TypeScript</li>
              <li>TanStack Router & Query</li>
              <li>TailwindCSS + ShadCN/UI</li>
              <li>Zustand + i18next</li>
              <li>WebSocket real-time</li>
            </ul>
          </div>
          <div>
            <h4 className="mb-2 text-sm font-medium text-green-600 dark:text-green-400">Backend</h4>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>Go 1.25 + Chi Router</li>
              <li>PostgreSQL + GORM</li>
              <li>Dragonfly (Redis-compatible)</li>
              <li>Multi-tenant orgs</li>
              <li>River job queue</li>
            </ul>
          </div>
          <div>
            <h4 className="mb-2 text-sm font-medium text-orange-600 dark:text-orange-400">DevOps</h4>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>Docker Compose</li>
              <li>Prometheus metrics</li>
              <li>Grafana dashboards</li>
              <li>GitHub Actions CI</li>
              <li>Swagger API docs</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}

// Reusable Feature Card
function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
      <div className="mb-2 flex items-center gap-2">
        <div className="text-blue-600 dark:text-blue-400">{icon}</div>
        <h4 className="font-medium text-gray-900 dark:text-white">{title}</h4>
      </div>
      <p className="text-sm text-gray-600 dark:text-gray-400">{description}</p>
    </div>
  );
}
