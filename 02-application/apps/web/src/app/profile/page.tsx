import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ProfileForm } from "@/components/profile/ProfileForm";
import { ChangePasswordForm } from "@/components/profile/ChangePasswordForm";
import { DeleteAccountForm } from "@/components/profile/DeleteAccountForm";

// Mock profile data - in real app, this would come from API
const mockProfile = {
  id: "1",
  email: "user@example.com",
  name: "John Doe",
  avatar: "",
  createdAt: "2024-01-01T00:00:00Z",
};

export default function ProfilePage() {
  return (
    <div className="container mx-auto max-w-4xl px-4 py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Account Settings</h1>
        <p className="text-muted-foreground">Manage your profile and account settings</p>
      </div>

      <Tabs defaultValue="profile" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="password">Password</TabsTrigger>
          <TabsTrigger value="danger">Danger Zone</TabsTrigger>
        </TabsList>
        <TabsContent value="profile">
          <ProfileForm profile={mockProfile} />
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
