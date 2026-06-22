import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TwoFASetup } from "@/components/security/TwoFASetup";
import { BiometricSettings } from "@/components/security/BiometricSettings";
import { WebAuthnSettings } from "@/components/auth/WebAuthnSettings";

export default function SecurityPage() {
  return (
    <div className="container mx-auto max-w-4xl px-4 py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Security Settings</h1>
        <p className="text-muted-foreground">
          Manage your account security and authentication methods
        </p>
      </div>

      <Tabs defaultValue="2fa" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="2fa">Two-Factor Auth</TabsTrigger>
          <TabsTrigger value="webauthn">Security Keys</TabsTrigger>
          <TabsTrigger value="biometric">Biometric</TabsTrigger>
        </TabsList>
        <TabsContent value="2fa">
          <TwoFASetup />
        </TabsContent>
        <TabsContent value="webauthn">
          <WebAuthnSettings />
        </TabsContent>
        <TabsContent value="biometric">
          <BiometricSettings />
        </TabsContent>
      </Tabs>
    </div>
  );
}