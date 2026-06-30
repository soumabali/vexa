"use client";

import React, { useState } from "react";
import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { MaterialIcon } from "@/components/ui/material-icon";
import type { Credential, CredentialType } from "./credential-list";

interface CredentialCardProps {
  credential: Credential;
  onUpdate: (id: string, data: Partial<Credential>) => void;
  onDelete: (id: string) => void;
  onFavorite: (id: string) => void;
  onClick: () => void;
  viewMode: "grid" | "list";
}

const typeTint: Record<CredentialType, string> = {
  "ssh-key": "bg-tertiary-container text-on-tertiary-container",
  password: "bg-primary-container text-on-primary-container",
  "api-key": "bg-secondary-container text-on-secondary-container",
  certificate: "bg-surface-container-high text-on-surface-variant",
  note: "bg-surface-container-high text-on-surface-variant",
};

const typeIcon: Record<CredentialType, string> = {
  "ssh-key": "key",
  password: "password",
  "api-key": "code",
  certificate: "verified",
  note: "note",
};

const typeLabels: Record<CredentialType, string> = {
  "ssh-key": "SSH Key",
  password: "Password",
  "api-key": "API Key",
  certificate: "Certificate",
  note: "Note",
};

export function CredentialCard({
  credential,
  onUpdate: _onUpdate,
  onDelete,
  onFavorite,
  onClick,
  viewMode,
}: CredentialCardProps) {
  const [isHovered, setIsHovered] = useState(false);

  const handleCopy = (text: string) => {
    navigator.clipboard?.writeText(text);
  };

  if (viewMode === "list") {
    return (
      <Card
        className="bg-surface-container border-outline-variant hover:border-primary cursor-pointer transition-all duration-200"
        onClick={onClick}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        <CardContent className="p-4 flex items-center gap-4">
          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${typeTint[credential.type]}`}>
            <MaterialIcon name={typeIcon[credential.type]} className="text-[24px]" />
          </div>
          <div className="flex items-center gap-3 flex-1">
            <Badge className={`${typeTint[credential.type]} border-0`}>
              {typeLabels[credential.type]}
            </Badge>
            <div className="flex-1">
              <h3 className="font-medium text-on-surface">{credential.name}</h3>
              {credential.host && (
                <p className="text-sm text-on-surface-variant">
                  {credential.username ? `${credential.username}@` : ""}
                  {credential.host}
                  {credential.port ? `:${credential.port}` : ""}
                </p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {credential.tags.map((tag) => (
              <Badge key={tag} variant="outline" className="text-xs border-outline-variant text-on-surface-variant">
                {tag}
              </Badge>
            ))}
          </div>

          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={(e) => {
                e.stopPropagation();
                onFavorite(credential.id);
              }}
            >
              <MaterialIcon
                name={credential.isFavorite ? "star" : "star_border"}
                size="sm"
                fill={credential.isFavorite}
                className={credential.isFavorite ? "text-warning" : "text-on-surface-variant"}
              />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={(e) => {
                e.stopPropagation();
              }}
            >
              <MaterialIcon name="share" size="sm" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0 hover:bg-error/10"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(credential.id);
              }}
            >
              <MaterialIcon name="delete" size="sm" className="text-error" />
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <TooltipProvider>
      <motion.div
        whileHover={{ scale: 1.02 }}
        whileTap={{ scale: 0.98 }}
      >
        <Card
          className="glass-card cursor-pointer transition-all duration-200 relative overflow-hidden"
          onClick={onClick}
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
        >
          {/* Favorite indicator */}
          {credential.isFavorite && (
            <div className="absolute top-2 right-2">
              <MaterialIcon name="star" size="sm" fill className="text-warning" />
            </div>
          )}

          <CardContent className="p-5 space-y-3">
            {/* Header */}
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <Badge className={`${typeTint[credential.type]} border-0`}>
                  {typeLabels[credential.type]}
                </Badge>
                <h3 className="text-body-lg font-bold text-on-surface truncate">{credential.name}</h3>
              </div>
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${typeTint[credential.type]}`}>
                <MaterialIcon name={typeIcon[credential.type]} className="text-[24px]" />
              </div>
            </div>

            {/* Host info */}
            {credential.host && (
              <div className="text-body-md text-on-surface-variant">
                <p className="font-mono-code">
                  {credential.username ? `${credential.username}@` : ""}
                  {credential.host}
                  {credential.port ? `:${credential.port}` : ""}
                </p>
              </div>
            )}

            {/* Tags */}
            {credential.tags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {credential.tags.slice(0, 3).map((tag) => (
                  <Badge key={tag} variant="outline" className="text-xs border-outline-variant text-on-surface-variant">
                    {tag}
                  </Badge>
                ))}
                {credential.tags.length > 3 && (
                  <Badge variant="outline" className="text-xs border-outline-variant text-on-surface-variant">
                    +{credential.tags.length - 3}
                  </Badge>
                )}
              </div>
            )}

            {/* Actions */}
            <div className="flex items-center justify-between pt-3 border-t border-outline-variant">
              <div className="flex items-center gap-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={(e) => {
                        e.stopPropagation();
                        onFavorite(credential.id);
                      }}
                    >
                      <MaterialIcon
                        name={credential.isFavorite ? "star" : "star_border"}
                        size="sm"
                        fill={credential.isFavorite}
                        className={credential.isFavorite ? "text-warning" : "text-on-surface-variant"}
                      />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    {credential.isFavorite ? "Remove favorite" : "Add favorite"}
                  </TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={(e) => {
                        e.stopPropagation();
                      }}
                    >
                      <MaterialIcon name="share" size="sm" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Share</TooltipContent>
                </Tooltip>
              </div>

              <div className={`flex items-center gap-1 transition-opacity ${isHovered ? "opacity-100" : "opacity-0"}`}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopy(
                          `${credential.username}@${credential.host}:${credential.port}`
                        );
                      }}
                    >
                      <MaterialIcon name="content_copy" size="sm" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Copy connection</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0 hover:bg-error/10"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(credential.id);
                      }}
                    >
                      <MaterialIcon name="delete" size="sm" className="text-error" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Delete</TooltipContent>
                </Tooltip>
              </div>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </TooltipProvider>
  );
}