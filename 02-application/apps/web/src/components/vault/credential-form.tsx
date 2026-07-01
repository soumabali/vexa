"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { MaterialIcon } from "@/components/ui/material-icon";
import type { Credential, CredentialType } from "./credential-list";

interface CredentialFormProps {
  credential?: Credential;
  onSubmit: (data: Omit<Credential, "id" | "createdAt" | "updatedAt">) => void;
  onCancel: () => void;
}

export function CredentialForm({ credential, onSubmit, onCancel }: CredentialFormProps) {
  const [type, setType] = useState<CredentialType>(credential?.type || "ssh-key");
  const [name, setName] = useState(credential?.name || "");
  const [username, setUsername] = useState(credential?.username || "");
  const [host, setHost] = useState(credential?.host || "");
  const [port, setPort] = useState(credential?.port?.toString() || "");
  const [folder, setFolder] = useState(credential?.folder || "Default");
  const [tags, setTags] = useState<string[]>(credential?.tags || []);
  const [tagInput, setTagInput] = useState("");
  const [privateKey, setPrivateKey] = useState("");
  const [password, setPassword] = useState("");
  const [note, setNote] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [certificate, setCertificate] = useState("");

  const handleAddTag = () => {
    if (tagInput.trim() && !tags.includes(tagInput.trim())) {
      setTags([...tags, tagInput.trim()]);
      setTagInput("");
    }
  };

  const handleRemoveTag = (tag: string) => {
    setTags(tags.filter((t) => t !== tag));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const data: Omit<Credential, "id" | "createdAt" | "updatedAt"> = {
      type,
      name,
      username: username || undefined,
      host: host || undefined,
      port: port ? parseInt(port) : undefined,
      folder,
      tags,
      isFavorite: credential?.isFavorite || false,
    };

    switch (type) {
      case "ssh-key":
        data.privateKey = privateKey;
        break;
      case "password":
        data.password = password;
        break;
      case "api-key":
        data.apiKey = apiKey;
        break;
      case "certificate":
        data.certificate = certificate;
        break;
      case "note":
        data.note = note;
        break;
    }

    onSubmit(data);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 bg-surface-container-high rounded-xl border border-outline-variant p-5">
      <Tabs value={type} onValueChange={(v) => setType(v as CredentialType)}>
        <TabsList className="grid grid-cols-5">
          <TabsTrigger value="ssh-key">SSH Key</TabsTrigger>
          <TabsTrigger value="password">Password</TabsTrigger>
          <TabsTrigger value="api-key">API Key</TabsTrigger>
          <TabsTrigger value="certificate">Certificate</TabsTrigger>
          <TabsTrigger value="note">Note</TabsTrigger>
        </TabsList>

        <div className="space-y-4 mt-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="name" className="text-on-surface">Name *</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Production Server"
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="folder" className="text-on-surface">Folder</Label>
              <Input
                id="folder"
                value={folder}
                onChange={(e) => setFolder(e.target.value)}
                placeholder="Default"
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="username" className="text-on-surface">Username</Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="e.g., root"
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="host" className="text-on-surface">Host</Label>
              <Input
                id="host"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                placeholder="e.g., 192.168.1.1"
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="port" className="text-on-surface">Port</Label>
            <Input
              id="port"
              type="number"
              value={port}
              onChange={(e) => setPort(e.target.value)}
              placeholder="22"
              className="bg-surface-container-lowest border-outline-variant text-on-surface"
            />
          </div>

          {/* Tags */}
          <div className="space-y-2">
            <Label className="text-on-surface">Tags</Label>
            <div className="flex items-center gap-2">
              <Input
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                placeholder="Add tag..."
                onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), handleAddTag())}
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
              <Button type="button" variant="outline" className="border-outline-variant" onClick={handleAddTag}>
                Add
              </Button>
            </div>
            <div className="flex flex-wrap gap-2 mt-2">
              {tags.map((tag) => (
                <Badge key={tag} variant="secondary" className="cursor-pointer" onClick={() => handleRemoveTag(tag)}>
                  {tag} ×
                </Badge>
              ))}
            </div>
          </div>

          {/* Type-specific fields */}
          <TabsContent value="ssh-key" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="privateKey" className="text-on-surface">Private Key</Label>
              <Textarea
                id="privateKey"
                value={privateKey}
                onChange={(e) => setPrivateKey(e.target.value)}
                placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                rows={6}
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </TabsContent>

          <TabsContent value="password" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="password" className="text-on-surface">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password"
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </TabsContent>

          <TabsContent value="api-key" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="apiKey" className="text-on-surface">API Key</Label>
              <Input
                id="apiKey"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter API key"
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </TabsContent>

          <TabsContent value="certificate" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="certificate" className="text-on-surface">Certificate</Label>
              <Textarea
                id="certificate"
                value={certificate}
                onChange={(e) => setCertificate(e.target.value)}
                placeholder="-----BEGIN CERTIFICATE-----"
                rows={6}
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </TabsContent>

          <TabsContent value="note" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="note" className="text-on-surface">Note</Label>
              <Textarea
                id="note"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Enter secure note..."
                rows={6}
                className="bg-surface-container-lowest border-outline-variant text-on-surface"
              />
            </div>
          </TabsContent>
        </div>
      </Tabs>

      <div className="flex justify-end gap-2 pt-4">
        <Button type="button" variant="outline" className="border-outline-variant" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" className="bg-primary text-on-primary hover:bg-primary/90">
          {credential ? "Update" : "Add"} Credential
        </Button>
      </div>
    </form>
  );
}
