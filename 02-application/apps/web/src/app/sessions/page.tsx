import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SessionList } from "@/components/sessions/SessionList";
import { LoginHistory } from "@/components/sessions/LoginHistory";

export default function SessionsPage() {
  return (
    <div className="container mx-auto max-w-4xl px-4 py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Session Management</h1>
        <p className="text-muted-foreground">
          Manage your active sessions and view login history
        </p>
      </div>

      <Tabs defaultValue="sessions" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="sessions">Active Sessions</TabsTrigger>
          <TabsTrigger value="history">Login History</TabsTrigger>
        </TabsList>
        <TabsContent value="sessions">
          <SessionList />
        </TabsContent>
        <TabsContent value="history">
          <LoginHistory />
        </TabsContent>
      </Tabs>
    </div>
  );
}
