import { useEffect, useRef, useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
  Bell,
  Camera,
  Github,
  Globe,
  Linkedin,
  Loader2,
  Lock,
  Mail,
  MapPin,
  Save,
  Settings,
  Trash2,
  Twitter,
  User,
} from "lucide-react";
import { toast } from "sonner";

import { Avatar, AvatarFallback, AvatarImage } from "../../components/ui/avatar";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Textarea } from "../../components/ui/textarea";
import { requireAuth } from "../../lib/guards";
import { AuthService } from "../../services/auth/authService";
import { SettingsService } from "../../services/settings/settingsService";
import type { SocialLinks } from "../../services/types";

export const Route = createFileRoute("/settings/profile")({
  beforeLoad: () => requireAuth(),
  component: ProfileSettingsPage,
});

function getUserInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

function parseSocialLinks(socialLinksStr?: string): SocialLinks {
  if (!socialLinksStr) return {};
  try {
    return JSON.parse(socialLinksStr);
  } catch {
    return {};
  }
}

function ProfileSettingsPage() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data: user, isLoading } = useQuery({
    queryKey: ["currentUser"],
    queryFn: () => AuthService.getCurrentUser(),
    staleTime: 60 * 1000,
  });

  const [formData, setFormData] = useState({
    name: "",
    email: "",
    bio: "",
    location: "",
    twitter: "",
    github: "",
    linkedin: "",
    website: "",
  });

  // Initialize form when user data loads
  useEffect(() => {
    if (user) {
      const socialLinks = parseSocialLinks(user.social_links);
      setFormData({
        name: user.name || "",
        email: user.email || "",
        bio: user.bio || "",
        location: user.location || "",
        twitter: socialLinks.twitter || "",
        github: socialLinks.github || "",
        linkedin: socialLinks.linkedin || "",
        website: socialLinks.website || "",
      });
    }
  }, [user]);

  const updateProfileMutation = useMutation({
    mutationFn: (data: { name?: string; email?: string; bio?: string; location?: string; social_links?: string }) =>
      SettingsService.updateProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      toast.success("Your profile has been updated successfully.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const uploadAvatarMutation = useMutation({
    mutationFn: (file: File) => SettingsService.uploadAvatar(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      toast.success("Your avatar has been uploaded successfully.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const deleteAvatarMutation = useMutation({
    mutationFn: () => SettingsService.deleteAvatar(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      toast.success("Your avatar has been removed.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const socialLinks: SocialLinks = {};
    if (formData.twitter) socialLinks.twitter = formData.twitter;
    if (formData.github) socialLinks.github = formData.github;
    if (formData.linkedin) socialLinks.linkedin = formData.linkedin;
    if (formData.website) socialLinks.website = formData.website;

    const updates: {
      name?: string;
      email?: string;
      bio?: string;
      location?: string;
      social_links?: string;
    } = {};

    if (formData.name !== user?.name) updates.name = formData.name;
    if (formData.email !== user?.email) updates.email = formData.email;
    if (formData.bio !== (user?.bio || "")) updates.bio = formData.bio;
    if (formData.location !== (user?.location || "")) updates.location = formData.location;

    const currentSocialLinks = parseSocialLinks(user?.social_links);
    const newSocialLinksStr = JSON.stringify(socialLinks);
    const currentSocialLinksStr = JSON.stringify(currentSocialLinks);
    if (newSocialLinksStr !== currentSocialLinksStr) {
      updates.social_links = newSocialLinksStr;
    }

    if (Object.keys(updates).length > 0) {
      updateProfileMutation.mutate(updates);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (!file.type.startsWith("image/")) {
        toast.error("Please upload an image file.");
        return;
      }
      if (file.size > 5 * 1024 * 1024) {
        toast.error("File size must be less than 5MB.");
        return;
      }
      uploadAvatarMutation.mutate(file);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-48 animate-pulse rounded-xl bg-gradient-to-r from-gray-200 to-gray-300" />
        <div className="h-64 animate-pulse rounded-lg bg-gray-200" />
      </div>
    );
  }

  const currentSocialLinks = parseSocialLinks(user?.social_links);
  const hasChanges =
    formData.name !== user?.name ||
    formData.email !== user?.email ||
    formData.bio !== (user?.bio || "") ||
    formData.location !== (user?.location || "") ||
    formData.twitter !== (currentSocialLinks.twitter || "") ||
    formData.github !== (currentSocialLinks.github || "") ||
    formData.linkedin !== (currentSocialLinks.linkedin || "") ||
    formData.website !== (currentSocialLinks.website || "");

  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <div className="relative overflow-hidden rounded-xl bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 p-8 shadow-lg">
        <div className="absolute inset-0 bg-black/10" />
        <div className="relative flex flex-col items-center gap-6 md:flex-row md:items-start">
          {/* Avatar with upload */}
          <div className="relative">
            <Avatar className="h-28 w-28 shadow-xl ring-4 ring-white/30">
              <AvatarImage
                src={user?.avatar_url || ""}
                alt={user?.name}
              />
              <AvatarFallback className="bg-white/20 text-3xl font-bold text-white">
                {getUserInitials(user?.name || "U")}
              </AvatarFallback>
            </Avatar>
            {uploadAvatarMutation.isPending && (
              <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/50">
                <Loader2 className="h-8 w-8 animate-spin text-white" />
              </div>
            )}
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="absolute right-0 bottom-0 rounded-full bg-white p-2 shadow-lg transition-transform hover:scale-110"
              disabled={uploadAvatarMutation.isPending}
            >
              <Camera className="h-4 w-4 text-gray-700" />
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              onChange={handleFileChange}
              className="hidden"
            />
          </div>

          {/* User Info */}
          <div className="flex-1 text-center md:text-left">
            <h1 className="text-3xl font-bold text-white">{user?.name}</h1>
            <p className="mt-1 text-blue-100">{user?.email}</p>
            <div className="mt-3 flex flex-wrap justify-center gap-2 md:justify-start">
              <Badge className="bg-white/20 text-white capitalize hover:bg-white/30">{user?.role || "User"}</Badge>
              {user?.email_verified && (
                <Badge className="bg-green-500/80 text-white hover:bg-green-500">Verified</Badge>
              )}
            </div>
            <p className="mt-3 text-sm text-blue-100">
              Member since{" "}
              {user?.created_at
                ? new Date(user.created_at).toLocaleDateString(undefined, {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })
                : "Unknown"}
            </p>
          </div>

          {/* Avatar Actions */}
          <div className="flex gap-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploadAvatarMutation.isPending}
              className="bg-white/20 text-white hover:bg-white/30"
            >
              <Camera className="mr-2 h-4 w-4" />
              Change Photo
            </Button>
            {user?.avatar_url && (
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => deleteAvatarMutation.mutate()}
                disabled={deleteAvatarMutation.isPending}
                className="bg-red-500/80 text-white hover:bg-red-500"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      </div>

      <form
        onSubmit={handleSubmit}
        className="space-y-6"
      >
        {/* Bio Section */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              About
            </CardTitle>
            <CardDescription>Tell others about yourself</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="bio">Bio</Label>
              <Textarea
                id="bio"
                placeholder="Write a short bio about yourself..."
                value={formData.bio}
                onChange={(e) => setFormData({ ...formData, bio: e.target.value })}
                rows={4}
                className="resize-none"
              />
              <p className="text-xs text-gray-500">Brief description for your profile. URLs are hyperlinked.</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="location">Location</Label>
              <div className="relative">
                <MapPin className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <Input
                  id="location"
                  placeholder="San Francisco, CA"
                  value={formData.location}
                  onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                  className="pl-10"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Personal Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Personal Information
            </CardTitle>
            <CardDescription>Update your name and email address</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="name">Full Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="Enter your name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email Address</Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  placeholder="Enter your email"
                />
                {formData.email !== user?.email && (
                  <p className="text-xs text-amber-600">Changing your email will require verification</p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Social Links */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              Social Links
            </CardTitle>
            <CardDescription>Connect your social profiles</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="twitter">Twitter / X</Label>
                <div className="relative">
                  <Twitter className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
                  <Input
                    id="twitter"
                    placeholder="username"
                    value={formData.twitter}
                    onChange={(e) => setFormData({ ...formData, twitter: e.target.value })}
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="github">GitHub</Label>
                <div className="relative">
                  <Github className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
                  <Input
                    id="github"
                    placeholder="username"
                    value={formData.github}
                    onChange={(e) => setFormData({ ...formData, github: e.target.value })}
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="linkedin">LinkedIn</Label>
                <div className="relative">
                  <Linkedin className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
                  <Input
                    id="linkedin"
                    placeholder="username"
                    value={formData.linkedin}
                    onChange={(e) => setFormData({ ...formData, linkedin: e.target.value })}
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="website">Website</Label>
                <div className="relative">
                  <Globe className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
                  <Input
                    id="website"
                    placeholder="https://example.com"
                    value={formData.website}
                    onChange={(e) => setFormData({ ...formData, website: e.target.value })}
                    className="pl-10"
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Save Button */}
        <div className="flex justify-end">
          <Button
            type="submit"
            disabled={!hasChanges || updateProfileMutation.isPending}
          >
            {updateProfileMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        </div>
      </form>

      {/* Account Status Grid */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="rounded-full bg-blue-100 p-3 dark:bg-blue-900">
                <Mail className="h-5 w-5 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Email Status</p>
                <Badge variant={user?.email_verified ? "default" : "secondary"}>
                  {user?.email_verified ? "Verified" : "Pending"}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="rounded-full bg-purple-100 p-3 dark:bg-purple-900">
                <User className="h-5 w-5 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Account Role</p>
                <Badge
                  variant="outline"
                  className="capitalize"
                >
                  {user?.role || "User"}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="rounded-full bg-green-100 p-3 dark:bg-green-900">
                <User className="h-5 w-5 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Account Status</p>
                <Badge variant={user?.is_active ? "default" : "destructive"}>
                  {user?.is_active ? "Active" : "Inactive"}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Settings Links */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Quick Settings
          </CardTitle>
          <CardDescription>Manage other aspects of your account</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-3">
            <Link
              to="/settings/security"
              className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <Lock className="h-5 w-5 text-gray-500" />
              <div>
                <p className="font-medium">Security</p>
                <p className="text-sm text-gray-500">Password & 2FA</p>
              </div>
            </Link>
            <Link
              to="/settings/preferences"
              className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <Settings className="h-5 w-5 text-gray-500" />
              <div>
                <p className="font-medium">Preferences</p>
                <p className="text-sm text-gray-500">Theme & Language</p>
              </div>
            </Link>
            <Link
              to="/settings/notifications"
              className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <Bell className="h-5 w-5 text-gray-500" />
              <div>
                <p className="font-medium">Notifications</p>
                <p className="text-sm text-gray-500">Email preferences</p>
              </div>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
