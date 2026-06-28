"use client";

import { useEffect, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { updateProfileSchema, changePasswordSchema, UpdateProfileInput, ChangePasswordInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import { AlertTriangle } from "lucide-react";
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
        <h1 className="text-3xl font-bold">Profile</h1>
        <p className="text-muted-foreground">Manage your profile and account settings</p>
      </div>

      <Tabs defaultValue="profile" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
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
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-center gap-4">
          <Avatar className="h-16 w-16">
            <AvatarImage src={profile.avatar} alt={profile.name} />
            <AvatarFallback className="text-lg">{initials}</AvatarFallback>
          </Avatar>
          <div>
            <CardTitle>Profile Information</CardTitle>
            <CardDescription>Update your profile information</CardDescription>
          </div>
        </div>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}
          {isSuccess && (
            <div className="p-3 bg-green-50 border border-green-200 rounded-md text-green-800 text-sm">
              Profile updated successfully!
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" placeholder="Your name" {...register("name")} />
            {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="you@example.com" {...register("email")} />
            {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
          </div>
          {profile.role && (
            <div className="space-y-2">
              <Label>Role</Label>
              <Input value={profile.role} disabled />
            </div>
          )}
          {profile.createdAt && (
            <div className="space-y-2">
              <Label>Member Since</Label>
              <Input value={new Date(profile.createdAt).toLocaleDateString()} disabled />
            </div>
          )}
          <Button type="submit" disabled={isLoading}>
            {isLoading ? <LoadingSpinner size="sm" /> : "Save Changes"}
          </Button>
        </CardContent>
      </form>
    </Card>
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
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Change Password</CardTitle>
        <CardDescription>Update your password to keep your account secure</CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}
          {isSuccess && (
            <div className="p-3 bg-green-50 border border-green-200 rounded-md text-green-800 text-sm">
              Password changed successfully!
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="currentPassword">Current Password</Label>
            <Input id="currentPassword" type="password" placeholder="••••••••" {...register("currentPassword")} />
            {errors.currentPassword && <p className="text-sm text-destructive">{errors.currentPassword.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="newPassword">New Password</Label>
            <Input id="newPassword" type="password" placeholder="••••••••" {...register("newPassword")} />
            {errors.newPassword && <p className="text-sm text-destructive">{errors.newPassword.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmNewPassword">Confirm New Password</Label>
            <Input id="confirmNewPassword" type="password" placeholder="••••••••" {...register("confirmNewPassword")} />
            {errors.confirmNewPassword && <p className="text-sm text-destructive">{errors.confirmNewPassword.message}</p>}
          </div>
          <Button type="submit" disabled={isLoading}>
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
      <Card className="w-full border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Account Deleted</CardTitle>
          <CardDescription>Your account has been permanently deleted.</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card className="w-full border-destructive">
      <CardHeader>
        <div className="flex items-center gap-2 text-destructive">
          <AlertTriangle className="h-5 w-5" />
          <CardTitle className="text-destructive">Delete Account</CardTitle>
        </div>
        <CardDescription>
          This action cannot be undone. All your data will be permanently removed.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}
          <div className="space-y-2">
            <Label htmlFor="deletePassword">Current Password</Label>
            <Input id="deletePassword" type="password" placeholder="••••••••" {...register("password")} />
            {errors.password && <p className="text-sm text-destructive">{errors.password.message}</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmDelete">Type DELETE to confirm</Label>
            <Input id="confirmDelete" placeholder="DELETE" {...register("confirmDelete")} />
            {errors.confirmDelete && <p className="text-sm text-destructive">{errors.confirmDelete.message}</p>}
          </div>
          <Button type="submit" variant="destructive" disabled={isLoading}>
            {isLoading ? <LoadingSpinner size="sm" /> : "Delete Account"}
          </Button>
        </CardContent>
      </form>
    </Card>
  );
}
