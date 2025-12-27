import { useState } from "react";

import { Link } from "@tanstack/react-router";
import { AnimatePresence, motion } from "framer-motion";
import {
  Activity,
  Bell,
  ChevronLeft,
  ChevronRight,
  Code,
  CreditCard,
  FileUp,
  Flag,
  Key,
  LayoutDashboard,
  Lock,
  Mail,
  ScrollText,
  Settings,
  Shield,
  Smartphone,
  Users,
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
    id: "files",
    title: "File Storage",
    icon: <FileUp className="h-5 w-5" />,
    description: "Upload files to S3 or database storage",
  },
  {
    id: "health",
    title: "System Health",
    icon: <Activity className="h-5 w-5" />,
    description: "Real-time health monitoring and status",
  },
  {
    id: "settings",
    title: "User Settings",
    icon: <Settings className="h-5 w-5" />,
    description: "Preferences, security, and account management",
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
              {currentStep === 1 && <BillingStep />}
              {currentStep === 2 && <AdminStep />}
              {currentStep === 3 && (
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
              {currentStep === 4 && (
                <HealthStep
                  healthStatus={healthStatus}
                  healthLoading={healthLoading}
                  testHealthCheck={testHealthCheck}
                />
              )}
              {currentStep === 5 && <SettingsStep />}
              {currentStep === 6 && <DeveloperStep />}
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

      <div className="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-2 font-semibold text-gray-900 dark:text-white">Monitored Components</h3>
        <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
          <li>Database connectivity and query performance</li>
          <li>Redis cache status and memory usage</li>
          <li>File storage availability</li>
          <li>Background job queue status</li>
        </ul>
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
          </ul>
        </div>
      </div>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h3 className="mb-4 font-semibold text-gray-900 dark:text-white">Tech Stack</h3>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <h4 className="mb-2 text-sm font-medium text-blue-600 dark:text-blue-400">Frontend</h4>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>React 19 + TypeScript</li>
              <li>TanStack Router & Query</li>
              <li>TailwindCSS + ShadCN/UI</li>
              <li>Vite + Hot Reload</li>
            </ul>
          </div>
          <div>
            <h4 className="mb-2 text-sm font-medium text-green-600 dark:text-green-400">Backend</h4>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>Go 1.25 + Chi Router</li>
              <li>PostgreSQL + GORM</li>
              <li>Redis Caching</li>
              <li>Docker + Docker Compose</li>
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
