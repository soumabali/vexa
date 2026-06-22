import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  UseFormRegister,
  FieldErrors,
  RegisterOptions,
  Path,
} from "react-hook-form";

interface FormFieldProps<T extends Record<string, unknown>> {
  label: string;
  name: Path<T>;
  type?: string;
  placeholder?: string;
  register: UseFormRegister<T>;
  errors: FieldErrors<T>;
  rules?: RegisterOptions<T, Path<T>>;
  disabled?: boolean;
}

export function FormField<T extends Record<string, unknown>>({
  label,
  name,
  type = "text",
  placeholder,
  register,
  errors,
  rules,
  disabled,
}: FormFieldProps<T>) {
  const errorMessage = errors[name]?.message as string | undefined;

  return (
    <div className="space-y-2">
      <Label htmlFor={name}>{label}</Label>
      <Input
        id={name}
        type={type}
        placeholder={placeholder}
        disabled={disabled}
        {...register(name, rules)}
        className={errorMessage ? "border-destructive" : ""}
      />
      {errorMessage && (
        <p className="text-sm text-destructive">{errorMessage}</p>
      )}
    </div>
  );
}
