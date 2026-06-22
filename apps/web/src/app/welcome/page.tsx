"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import Link from "next/link";
import { useState } from "react";
import { CheckCircle2, Shield, Key, Globe, ArrowRight } from "lucide-react";

const onboardingSteps = [
  {
    title: "Welcome to vexa",
    description: "Self-hosted SSH access management",
    icon: Shield,
  },
  {
    title: "Add Your First SSH Key",
    description: "Upload or generate SSH keys to start connecting to your servers.",
    icon: Key,
  },
  {
    title: "Connect to a Server",
    description: "Add server details and connect using your SSH keys.",
    icon: Globe,
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

  const CurrentIcon = onboardingSteps[currentStep].icon;

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12">
      <Card className="w-full max-w-lg">
        <CardHeader>
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
            <span className="text-sm text-muted-foreground">
              {currentStep + 1} of {onboardingSteps.length}
            </span>
          </div>
          <Progress value={progress} className="mb-4" />
          <div className="flex items-center gap-3">
            <div className="p-3 bg-primary/10 rounded-full">
              <CurrentIcon className="h-6 w-6 text-primary" />
            </div>
            <div>
              <CardTitle>{onboardingSteps[currentStep].title}</CardTitle>
              <CardDescription>{onboardingSteps[currentStep].description}</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {currentStep === onboardingSteps.length - 1 && (
            <div className="space-y-4">
              <div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
                <CheckCircle2 className="h-5 w-5 text-green-500" />
                <span className="text-sm">Email verified</span>
              </div>
              <div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
                <CheckCircle2 className="h-5 w-5 text-green-500" />
                <span className="text-sm">Account created</span>
              </div>
            </div>
          )}
        </CardContent>
        <CardFooter className="flex justify-between">
          <Button variant="ghost" onClick={handleSkip}>
            Skip for now
          </Button>
          {currentStep < onboardingSteps.length - 1 ? (
            <Button onClick={handleNext}>
              Next
              <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          ) : (
            <Link href="/dashboard">
              <Button>
                Get Started
                <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            </Link>
          )}
        </CardFooter>
      </Card>
    </div>
  );
}
