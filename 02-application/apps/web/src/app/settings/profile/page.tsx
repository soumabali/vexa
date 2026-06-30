"use client";

import { useEffect, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { updateProfileSchema, changePasswordSchema, UpdateProfileInput, ChangePasswordInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import { MaterialIcon } from "@/components/ui/material-icon";
import { z } from "zod";

interface ProfileData {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  createdAt?: string;
  role?: string;
  mfa_verified?: boolean;
}

export default function ProfileSettingsPage() {
  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function fetchProfile() {
      try {
        const data = (await authApi.getUserProfile().catch(() => authApi.getProfile())) as {
          id?: string;
          email?: string;
          name?: string;
          avatar?: string;
          createdAt?: string;
          role?: string;
          mfa_verified?: boolean;
        };
        setProfile({
          id: data.id || "",
          email: data.email || "",
          name: data.name || "",
          avatar: data.avatar || "",
          createdAt: data.createdAt,
          role: data.role || "",
          mfa_verified: data.mfa_verified,
        });
      } catch (err) {
        setProfile({
          id: "",
          email: "",
          name: "",
          avatar: "",
          createdAt: undefined,
          role: "",
          mfa_verified: false,
        });
      } finally {
        setLoading(false);
      }
    }
    fetchProfile();
  }, []);

  if (loading) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-12 flex justify-center">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-12">
        <ErrorDisplay message={error} />
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-12">
        <ErrorDisplay message="Profile not found" />
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-4xl px-4 py-12">
      <div className="mb-8">
        <h1 className="text-headline-lg text-on-surface mb-2">Profile Settings</h1>
        <p className="text-body-lg text-on-surface-variant">
          Manage your profile and account settings
        </p>
      </div>

      <Tabs defaultValue="profile" className="w-full">
        <TabsList className="grid w-full grid-cols-3 bg-surface-container border border-outline-variant">
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="password">Password</TabsTrigger>
          <TabsTrigger value="danger">Danger Zone</TabsTrigger>
        </TabsList>
        <TabsContent value="profile">
          <ProfileForm profile={profile} onUpdate={setProfile} />
        </TabsContent>
        <TabsContent value="password">
          <ChangePasswordForm />
        </TabsContent>
        <TabsContent value="danger">
          <DeleteAccountForm />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function ProfileForm({ profile, onUpdate }: { profile: ProfileData; onUpdate: (p: ProfileData) => void }) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<UpdateProfileInput>({
    resolver: zodResolver(updateProfileSchema),
    defaultValues: {
      name: profile.name,
      email: profile.email,
    },
  });

  const onSubmit = async (data: UpdateProfileInput) => {
    setIsLoading(true);
    setError("");
    setIsSuccess(false);
    try {
      await authApi.updateUserProfile(data);
      onUpdate({ ...profile, ...data });
      setIsSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update profile");
    } finally {
      setIsLoading(false);
    }
  };

  const initials = (profile.name || profile.email || "U")
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);

  return (
    <div className="space-y-6">
      {/* Avatar card */}
      <Card className="bg-surface-container border border-outline-variant rounded-xl">
        <CardHeader>
          <CardTitle className="text-on-surface">Avatar</CardTitle>
          <CardDescription className="text-on-surface-variant">
            Upload a profile photo or use your initials
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-6">
            <Avatar className="w-20 h-20 rounded-full bg-secondary-container border-2 border-outline-variant">
              <AvatarImage src={profile.avatar} alt={profile.name} />
              <AvatarFallback className="text-lg text-on-secondary-container bg-secondary-container">
                {initials}
              </AvatarFallback>
            </Avatar>
            <div className="flex gap-3">
              <Button type="button" className="bg-primary text-on-primary hover:bg-primary/90">
                <MaterialIcon name="upload" size="sm" className="mr-2" />
                Upload
              </Button>
              <Button type="button" variant="outline" className="border-outline-variant text-on-surface">
                <MaterialIcon name="delete" size="sm" className="mr-2" />
                Remove
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Identity card */}
      <Card className="bg-surface-container border border-outline-variant rounded-xl">
        <CardHeader>
          <CardTitle className="text-on-surface">Identity</CardTitle>
          <CardDescription className="text-on-surface-variant">
            Your public identity on this instance
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit(onSubmit)}>
          <CardContent className="space-y-4">
            {error && <ErrorDisplay message={error} />}
            {isSuccess && (
              <div className="p-3 bg-success/10 border border-success/30 rounded-md text-success text-sm flex items-center gap-2">
                <MaterialIcon name="check_circle" size="sm" />
                Profile updated successfully!
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="name" className="text-on-surface">Full name</Label>
              <Input
                id="name"
                placeholder="Your name"
                className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
                {...register("name")}
              />
              {errors.name && <p className="text-sm text-error">{errors.name.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="email" className="text-on-surface">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
                {...register("email")}
              />
              {errors.email && <p className="text-sm text-error">{errors.email.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="bio" className="text-on-surface">Bio</Label>
              <Input
                id="bio"
                placeholder="Short description"
                className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              />
            </div>
            {profile.role && (
              <div className="space-y-2">
                <Label className="text-on-surface">Role</Label>
                <Input
                  value={profile.role}
                  disabled
                  className="bg-surface-container-lowest border-outline-variant text-on-surface-variant"
                />
              </div>
            )}
            {profile.createdAt && (
              <div className="space-y-2">
                <Label className="text-on-surface">Member Since</Label>
                <Input
                  value={new Date(profile.createdAt).toLocaleDateString()}
                  disabled
                  className="bg-surface-container-lowest border-outline-variant text-on-surface-variant"
                />
              </div>
            )}
            <Button type="submit" disabled={isLoading} className="bg-primary text-on-primary hover:bg-primary/90">
              {isLoading ? <LoadingSpinner size="sm" /> : "Save Changes"}
            </Button>
          </CardContent>
        </form>
      </Card>

      {/* Public profile card */}
      <Card className="bg-surface-container border border-outline-variant rounded-xl">
        <CardHeader>
          <CardTitle className="text-on-surface">Public profile</CardTitle>
          <CardDescription className="text-on-surface-variant">
            Control how your profile appears to others
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-on-surface">Visibility</Label>
              <p className="text-sm text-on-surface-variant">Make your profile visible to other users</p>
            </div>
            <Switch />
          </div>
          <div className="space-y-2">
            <Label className="text-on-surface">Public URL</Label>
            <div className="flex items-center gap-2">
              <Input
                readOnly
                value={`https://vexa.local/u/${profile.id || "profile"}`}
                className="bg-surface-container-lowest border-outline-variant text-on-surface-variant font-mono text-sm"
              />
              <Button type="button" variant="outline" className="border-outline-variant text-on-surface">
                <MaterialIcon name="content_copy" size="sm" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function ChangePasswordForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ChangePasswordInput>({
    resolver: zodResolver(changePasswordSchema),
  });

  const onSubmit = async (data: ChangePasswordInput) => {
    setIsLoading(true);
    setError("");
    setIsSuccess(false);
    try {
      await authApi.changePassword({
        currentPassword: data.currentPassword,
        newPassword: data.newPassword,
      });
      setIsSuccess(true);
      reset();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to change password");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card className="bg-surface-container border border-outline-variant rounded-xl">
      <CardHeader>
        <CardTitle className="text-on-surface">Change Password</CardTitle>
        <CardDescription className="text-on-surface-variant">
          Update your password to keep your account secure
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}
          {isSuccess && (
            <div className="p-3 bg-success/10 border border-success/30 rounded-md text-success text-sm flex items-center gap-2">
              <MaterialIcon name="check_circle" size="sm" />
              Password changed successfully!
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="currentPassword" className="text-on-surface">Current Password</Label>
            <Input
              id="currentPassword"
              type="password"
              placeholder="••••••••"
              className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("currentPassword")}
            />
            {errors.currentPassword && <p className="text-sm text-error">{errors.currentPassword.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="newPassword" className="text-on-surface">New Password</Label>
            <Input
              id="newPassword"
              type="password"
              placeholder="••••••••"
              className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("newPassword")}
            />
            {errors.newPassword && <p className="text-sm text-error">{errors.newPassword.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmNewPassword" className="text-on-surface">Confirm New Password</Label>
            <Input
              id="confirmNewPassword"
              type="password"
              placeholder="••••••••"
              className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("confirmNewPassword")}
            />
            {errors.confirmNewPassword && <p className="text-sm text-error">{errors.confirmNewPassword.message}</p>}
          </div>
          <Button type="submit" disabled={isLoading} className="bg-primary text-on-primary hover:bg-primary/90">
            {isLoading ? <LoadingSpinner size="sm" /> : "Change Password"}
          </Button>
        </CardContent>
      </form>
    </Card>
  );
}

const deleteAccountSchema = z.object({
  password: z.string().min(1, "Password is required"),
  confirmDelete: z.enum(["DELETE"]),
});

type DeleteAccountInput = z.output<typeof deleteAccountSchema>;

function DeleteAccountForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DeleteAccountInput>({
    resolver: zodResolver(deleteAccountSchema),
  });

  const onSubmit = async (data: DeleteAccountInput) => {
    setIsLoading(true);
    setError("");
    try {
      await authApi.deleteAccount(data.password);
      setIsSuccess(true);
      reset();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete account");
    } finally {
      setIsLoading(false);
    }
  };

  if (isSuccess) {
    return (
      <Card className="bg-surface-container border border-error rounded-xl">
        <CardHeader>
          <CardTitle className="text-error">Account Deleted</CardTitle>
          <CardDescription className="text-on-surface-variant">
            Your account has been permanently deleted.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card className="bg-surface-container border border-error rounded-xl">
      <CardHeader>
        <div className="flex items-center gap-2 text-error">
          <MaterialIcon name="warning" className="text-error" />
          <CardTitle className="text-error">Delete Account</CardTitle>
        </div>
        <CardDescription className="text-on-surface-variant">
          This action cannot be undone. All your data will be permanently removed.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}
          <div className="space-y-2">
            <Label htmlFor="deletePassword" className="text-on-surface">Current Password</Label>
            <Input
              id="deletePassword"
              type="password"
              placeholder="••••••••"
              className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("password")}
            />
            {errors.password && <p className="text-sm text-error">{errors.password.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmDelete" className="text-on-surface">Type DELETE to confirm</Label>
            <Input
              id="confirmDelete"
              placeholder="DELETE"
              className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("confirmDelete")}
            />
            {errors.confirmDelete && <p className="text-sm text-error">{errors.confirmDelete.message}</p>}
          </div>
          <Button type="submit" variant="destructive" disabled={isLoading} className="bg-error text-on-error hover:bg-error/90">
            {isLoading ? <LoadingSpinner size="sm" /> : "Delete Account"}
          </Button>
        </CardContent>
      </form>
    </Card>
  );
}