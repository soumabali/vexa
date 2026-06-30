"use client";

import React, { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  createHostSchema,
  updateHostSchema,
  HostType,
  HostResponse,
} from "@/lib/validations/hosts";
import { hostsApi } from "@/lib/api/hosts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import { cn } from "@/lib/utils";
import { MaterialIcon } from "@/components/ui/material-icon";

interface HostFormProps {
  host?: HostResponse;
  onSuccess?: (host: HostResponse) => void;
  onCancel?: () => void;
}

const hostTypes = [
  { value: "ssh", label: "SSH", icon: "dns" },
  { value: "rdp", label: "RDP", icon: "computer" },
  { value: "vnc", label: "VNC", icon: "computer" },
] as const;

const colorPresets = [
  "#EF4444", "#F97316", "#EAB308", "#22C55E",
  "#14B8A6", "#3B82F6", "#8B5CF6", "#EC4899",
];

const tagColors = [
  "bg-error/10 text-error border-error/20",
  "bg-tertiary-container text-on-tertiary-container border-outline-variant",
  "bg-primary-container text-on-primary-container border-outline-variant",
  "bg-secondary-container text-on-secondary-container border-outline-variant",
  "bg-primary-container text-on-primary-container border-outline-variant",
  "bg-tertiary-container text-on-tertiary-container border-outline-variant",
];

export function HostForm({ host, onSuccess, onCancel }: HostFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [selectedHostType, setSelectedHostType] = useState<HostType>(
    (host?.hostType as HostType) || "ssh"
  );
  const [tags, setTags] = useState<string[]>(host?.tags || []);
  const [tagInput, setTagInput] = useState("");
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [selectedColor, setSelectedColor] = useState(
    host?.color || colorPresets[5]
  );

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(host ? updateHostSchema : createHostSchema),
    defaultValues: host
      ? {
          name: host.name,
          hostType: host.hostType as HostType,
          host: host.host || host.address || "",
          port: host.port,
          username: host.username || "",
          description: host.description || "",
          groupId: host.groupId || undefined,
          tags: host.tags || [],
          color: host.color || undefined,
          favorite: host.favorite,
          sshOptions: host.sshOptions || undefined,
          rdpOptions: host.rdpOptions || undefined,
          vncOptions: host.vncOptions || undefined,
        }
      : {
          hostType: "ssh",
          port: 22,
          tags: [],
          favorite: false,
        },
  });

  const watchedValues = watch();

  const handleAddTag = () => {
    const trimmed = tagInput.trim();
    if (trimmed && !tags.includes(trimmed)) {
      setTags([...tags, trimmed]);
      setValue("tags", [...tags, trimmed]);
    }
    setTagInput("");
  };

  const handleRemoveTag = (tag: string) => {
    setTags(tags.filter((t) => t !== tag));
    setValue(
      "tags",
      tags.filter((t) => t !== tag)
    );
  };

  const handleTagKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      handleAddTag();
    }
  };

  const onSubmit = async (data: unknown) => {
    setIsLoading(true);
    setError("");
    try {
      const formData = data as {
        name?: string;
        host?: string;
        port?: number;
        username?: string;
        description?: string;
        groupId?: string;
        tags?: string[];
        color?: string;
        favorite?: boolean;
        sshOptions?: Record<string, unknown>;
        rdpOptions?: Record<string, unknown>;
        vncOptions?: Record<string, unknown>;
      };
      const payload = { ...formData, hostType: selectedHostType, type: selectedHostType };
      let result: HostResponse;
      if (host) {
        result = await hostsApi.update(host.id, payload);
      } else {
        result = await hostsApi.create(payload);
      }
      onSuccess?.(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save host");
    } finally {
      setIsLoading(false);
    }
  };

  return (
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="space-y-6"
      >
      {error && <ErrorDisplay message={error} />}

      {/* Host Type Selector */}
      <div className="space-y-2">
        <Label className="text-label-md text-on-surface-variant">Connection Type</Label>
        <div className="grid grid-cols-3 gap-2">
          {hostTypes.map(({ value, label, icon }) => (
            <button
              key={value}
              type="button"
              onClick={() => {
                console.log("HostForm set hostType", value);
                setSelectedHostType(value as HostType);
                setValue("hostType", value as HostType);
              }}
              className={cn(
                "flex flex-col items-center gap-2 rounded-lg border-2 p-3 transition-all bg-surface-container-low",
                selectedHostType === value
                  ? "border-primary bg-primary-container text-on-primary-container"
                  : "border-outline-variant text-on-surface-variant hover:border-primary"
              )}
            >
              <MaterialIcon name={icon} size="md" />
              <span className="text-label-md">{label}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Basic Info */}
      <div className="grid grid-cols-2 gap-4">
        <div className="col-span-2 space-y-2">
          <Label htmlFor="name" className="text-label-md text-on-surface-variant">Name *</Label>
          <Input
            id="name"
            placeholder="Production Server"
            {...register("name")}
            className={cn("bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary", errors.name && "border-error")}
          />
          {errors.name && (
            <p className="text-sm text-error">{errors.name.message as string}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="host" className="text-label-md text-on-surface-variant">Host / IP *</Label>
          <Input
            id="host"
            placeholder="192.168.1.100"
            {...register("host")}
            className={cn("bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary", errors.host && "border-error")}
          />
          {errors.host && (
            <p className="text-sm text-error">{errors.host.message as string}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="port" className="text-label-md text-on-surface-variant">Port *</Label>
          <Input
            id="port"
            type="number"
            placeholder={selectedHostType === "ssh" ? "22" : selectedHostType === "rdp" ? "3389" : "5900"}
            {...register("port", { valueAsNumber: true })}
            className={cn("bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary", errors.port && "border-error")}
          />
          {errors.port && (
            <p className="text-sm text-error">{errors.port.message as string}</p>
          )}
        </div>

        <div className="col-span-2 space-y-2">
          <Label htmlFor="username" className="text-label-md text-on-surface-variant">Username</Label>
          <Input
            id="username"
            placeholder="root"
            {...register("username")}
            className="bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
          />
        </div>

        <div className="col-span-2 space-y-2">
          <Label htmlFor="description" className="text-label-md text-on-surface-variant">Description</Label>
          <textarea
            id="description"
            placeholder="Brief description of this host..."
            rows={2}
            className={cn(
              "flex w-full rounded-lg border border-outline-variant bg-surface-container-low px-3 py-2 text-body-md placeholder:text-on-surface-variant focus-visible:outline-none focus:border-primary disabled:cursor-not-allowed disabled:opacity-50"
            )}
            {...register("description")}
          />
        </div>
      </div>

      {/* Color Picker */}
      <div className="space-y-2">
        <Label className="text-label-md text-on-surface-variant">Color Label</Label>
        <div className="flex items-center gap-2">
          <div
            className="h-6 w-6 rounded border border-outline-variant"
            style={{ backgroundColor: selectedColor }}
          />
          <div className="flex gap-1">
            {colorPresets.map((color) => (
              <button
                key={color}
                type="button"
                onClick={() => {
                  setSelectedColor(color);
                  setValue("color", color);
                }}
                className={cn(
                  "h-5 w-5 rounded border-2 transition-all",
                  selectedColor === color ? "border-primary scale-110" : "border-outline-variant"
                )}
                style={{ backgroundColor: color }}
              />
            ))}
          </div>
          <Input
            type="color"
            value={selectedColor}
            onChange={(e) => {
              setSelectedColor(e.target.value);
              setValue("color", e.target.value);
            }}
            className="h-8 w-16 cursor-pointer bg-surface-container-low border border-outline-variant rounded-lg"
          />
        </div>
      </div>

      {/* Tags */}
      <div className="space-y-2">
        <Label className="text-label-md text-on-surface-variant">Tags</Label>
        <div className="flex flex-wrap gap-2 mb-2">
          {tags.map((tag, i) => (
            <Badge key={tag} className={tagColors[i % tagColors.length]}>
              {tag}
              <button
                type="button"
                onClick={() => handleRemoveTag(tag)}
                className="ml-1 hover:text-error"
              >
                <MaterialIcon name="close" size="sm" />
              </button>
            </Badge>
          ))}
        </div>
        <div className="flex gap-2">
          <Input
            value={tagInput}
            onChange={(e) => setTagInput(e.target.value)}
            onKeyDown={handleTagKeyDown}
            placeholder="Add tag and press Enter"
            className="flex-1 bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
          />
          <Button type="button" variant="outline" onClick={handleAddTag} className="border border-outline-variant text-on-surface">
            <MaterialIcon name="add" size="sm" />
          </Button>
        </div>
      </div>

      {/* SSH Options */}
      {selectedHostType === "ssh" && (
        <div className="space-y-4 rounded-lg border border-outline-variant bg-surface-container-low p-4">
          <div className="flex items-center justify-between">
            <Label className="text-label-md text-on-surface-variant">SSH Options</Label>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="text-on-surface-variant"
            >
              {showAdvanced ? <MaterialIcon name="expand_less" size="sm" /> : <MaterialIcon name="expand_more" size="sm" />}
              {showAdvanced ? "Hide" : "Show"} Options
            </Button>
          </div>

          {showAdvanced && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="ssh.identityFile" className="text-label-md text-on-surface-variant">Identity File</Label>
                  <Input
                    id="ssh.identityFile"
                    placeholder="~/.ssh/id_ed25519"
                    {...register("sshOptions.identityFile")}
                    className="bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="ssh.jumpHost" className="text-label-md text-on-surface-variant">Jump Host</Label>
                  <Input
                    id="ssh.jumpHost"
                    placeholder="bastion@example.com"
                    {...register("sshOptions.jumpHost")}
                    className="bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
                  />
                </div>
              </div>
              <div className="flex flex-wrap gap-4">
                <div className="flex items-center gap-2">
                  <Switch
                    id="ssh.keepAlive"
                    checked={watchedValues.sshOptions?.keepAlive ?? true}
                    onCheckedChange={(v) =>
                      setValue("sshOptions.keepAlive", v as boolean)
                    }
                  />
                  <Label htmlFor="ssh.keepAlive" className="text-label-md text-on-surface-variant">Keep Alive</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    id="ssh.compress"
                    checked={watchedValues.sshOptions?.compress ?? false}
                    onCheckedChange={(v) =>
                      setValue("sshOptions.compress", v as boolean)
                    }
                  />
                  <Label htmlFor="ssh.compress" className="text-label-md text-on-surface-variant">Compress</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    id="ssh.forwardingAgent"
                    checked={watchedValues.sshOptions?.forwardingAgent ?? false}
                    onCheckedChange={(v) =>
                      setValue("sshOptions.forwardingAgent", v as boolean)
                    }
                  />
                  <Label htmlFor="ssh.forwardingAgent" className="text-label-md text-on-surface-variant">Forward Agent</Label>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* RDP Options */}
      {selectedHostType === "rdp" && (
        <div className="space-y-4 rounded-lg border border-outline-variant bg-surface-container-low p-4">
          <Label className="text-label-md text-on-surface-variant">RDP Options</Label>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rdp.domain" className="text-label-md text-on-surface-variant">Domain</Label>
              <Input
                id="rdp.domain"
                placeholder="WORKGROUP"
                {...register("rdpOptions.domain")}
                className="bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
              />
            </div>
            <div className="space-y-2">
              <Label className="text-label-md text-on-surface-variant">Color Depth</Label>
              <Select
                defaultValue={(host?.rdpOptions?.colorDepth as string) || "32"}
                onValueChange={(v) => setValue("rdpOptions.colorDepth", v as "16" | "24" | "32")}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="16">16-bit</SelectItem>
                  <SelectItem value="24">24-bit</SelectItem>
                  <SelectItem value="32">32-bit</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rdp.width" className="text-label-md text-on-surface-variant">Width</Label>
              <Input
                id="rdp.width"
                type="number"
                defaultValue={host?.rdpOptions?.width as number ?? 1920}
                {...register("rdpOptions.width", { valueAsNumber: true })}
                className="bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rdp.height" className="text-label-md text-on-surface-variant">Height</Label>
              <Input
                id="rdp.height"
                type="number"
                defaultValue={host?.rdpOptions?.height as number ?? 1080}
                {...register("rdpOptions.height", { valueAsNumber: true })}
                className="bg-surface-container-low border border-outline-variant rounded-lg focus:border-primary"
              />
            </div>
          </div>
          <div className="flex flex-wrap gap-4">
            <div className="flex items-center gap-2">
              <Switch
                id="rdp.clipboardRedirect"
                defaultChecked={(host?.rdpOptions?.clipboardRedirect as boolean) ?? true}
                onCheckedChange={(v) => setValue("rdpOptions.clipboardRedirect", v)}
              />
              <Label htmlFor="rdp.clipboardRedirect" className="text-sm">Clipboard</Label>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="rdp.driveRedirect"
                defaultChecked={(host?.rdpOptions?.driveRedirect as boolean) ?? false}
                onCheckedChange={(v) => setValue("rdpOptions.driveRedirect", v)}
              />
              <Label htmlFor="rdp.driveRedirect" className="text-label-md text-on-surface-variant">Drive Redirect</Label>
            </div>
          </div>
        </div>
      )}

      {/* VNC Options */}
      {selectedHostType === "vnc" && (
        <div className="space-y-4 rounded-lg border border-outline-variant bg-surface-container-low p-4">
          <Label className="text-label-md text-on-surface-variant">VNC Options</Label>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label className="text-label-md text-on-surface-variant">Quality</Label>
              <Select
                defaultValue={(host?.vncOptions?.quality as string) || "auto"}
                onValueChange={(v) =>
                  setValue("vncOptions.quality", v as "low" | "medium" | "high" | "auto")
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="auto">Auto</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label className="text-label-md text-on-surface-variant">Compression</Label>
              <Select
                defaultValue={(host?.vncOptions?.compression as string) || "auto"}
                onValueChange={(v) =>
                  setValue("vncOptions.compression", v as "low" | "medium" | "high" | "auto")
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="auto">Auto</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex flex-wrap gap-4">
            <div className="flex items-center gap-2">
              <Switch
                id="vnc.viewOnly"
                defaultChecked={(host?.vncOptions?.viewOnly as boolean) ?? false}
                onCheckedChange={(v) => setValue("vncOptions.viewOnly", v)}
              />
              <Label htmlFor="vnc.viewOnly" className="text-label-md text-on-surface-variant">View Only</Label>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="vnc.shared"
                defaultChecked={(host?.vncOptions?.shared as boolean) ?? true}
                onCheckedChange={(v) => setValue("vncOptions.shared", v)}
              />
              <Label htmlFor="vnc.shared" className="text-label-md text-on-surface-variant">Shared Mode</Label>
            </div>
          </div>
        </div>
      )}

      {/* Favorite toggle */}
      <div className="flex items-center gap-2">
        <Switch
          id="favorite"
          checked={watchedValues.favorite ?? false}
          onCheckedChange={(v) => setValue("favorite", v)}
        />
        <Label htmlFor="favorite" className="text-label-md text-on-surface-variant">Add to favorites</Label>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end gap-3 border-t border-outline-variant pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} className="border border-outline-variant text-on-surface rounded-lg">
            Cancel
          </Button>
        )}
        <Button type="submit" disabled={isLoading} className="bg-primary text-on-primary hover:opacity-90 rounded-lg">
          {isLoading ? (
            <>
              <LoadingSpinner className="h-4 w-4" />
              {host ? "Updating..." : "Creating..."}
            </>
          ) : (
            host ? "Update Host" : "Create Host"
          )}
        </Button>
      </div>
    </form>
  );
}