"use client";

import { Button } from "@/components/ui/button";
import { MaterialIcon } from "@/components/ui/material-icon";
import { Progress } from "@/components/ui/progress";
import Link from "next/link";
import { useState } from "react";

const onboardingSteps = [
  {
    title: "Welcome to vexa",
    description: "Self-hosted SSH access management",
    icon: "shield",
  },
  {
    title: "Add Your First SSH Key",
    description: "Upload or generate SSH keys to start connecting to your servers.",
    icon: "key",
  },
  {
    title: "Connect to a Server",
    description: "Add server details and connect using your SSH keys.",
    icon: "public",
  },
];

export default function WelcomePage() {
  const [currentStep, setCurrentStep] = useState(0);

  const progress = ((currentStep + 1) / onboardingSteps.length) * 100;

  const handleNext = () => {
    if (currentStep < onboardingSteps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleSkip = () => {
    window.location.href = "/dashboard";
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="bg-surface-container border border-outline-variant rounded-xl p-8 max-w-lg w-full">
        <div className="flex items-center gap-3 mb-8">
          <div className="bg-primary rounded-lg w-12 h-12 flex items-center justify-center">
            <MaterialIcon name="terminal" className="text-on-primary" fill />
          </div>
          <span className="text-2xl font-semibold text-on-surface">Vexa</span>
        </div>

        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            {onboardingSteps.map((_, index) => (
              <div
                key={index}
                className={`w-2 h-2 rounded-full ${
                  index <= currentStep ? "bg-primary" : "bg-muted"
                }`}
              />
            ))}
          </div>
          <span className="text-sm text-on-surface-variant">
            {currentStep + 1} of {onboardingSteps.length}
          </span>
        </div>

        <Progress value={progress} className="mb-6" />

        <div className="flex items-center gap-3 mb-6">
          <div className="p-3 bg-primary/10 rounded-full">
            <MaterialIcon
              name={onboardingSteps[currentStep].icon}
              className="text-primary"
              size="lg"
            />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-on-surface">
              {onboardingSteps[currentStep].title}
            </h2>
            <p className="text-sm text-on-surface-variant">
              {onboardingSteps[currentStep].description}
            </p>
          </div>
        </div>

        {currentStep === onboardingSteps.length - 1 && (
          <div className="space-y-3 mb-6">
            <div className="flex items-center gap-3 p-3 bg-surface-container-high rounded-lg">
              <MaterialIcon name="check_circle" className="text-primary" fill size="sm" />
              <span className="text-sm text-on-surface">Email verified</span>
            </div>
            <div className="flex items-center gap-3 p-3 bg-surface-container-high rounded-lg">
              <MaterialIcon name="check_circle" className="text-primary" fill size="sm" />
              <span className="text-sm text-on-surface">Account created</span>
            </div>
          </div>
        )}

        <div className="flex justify-between items-center">
          <Button variant="ghost" onClick={handleSkip}>
            Skip for now
          </Button>
          {currentStep < onboardingSteps.length - 1 ? (
            <Button onClick={handleNext}>
              Next
              <MaterialIcon name="arrow_right" className="ml-2" size="sm" />
            </Button>
          ) : (
            <Link href="/dashboard">
              <Button className="bg-primary text-on-primary">
                Get Started
                <MaterialIcon name="arrow_right" className="ml-2 text-on-primary" size="sm" />
              </Button>
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}