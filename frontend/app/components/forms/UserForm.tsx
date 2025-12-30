import React, { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useTranslation } from "react-i18next";

import type { User } from "../../services";

interface UserFormProps {
  onSubmit: (name: string, email: string, id?: number) => void;
  initialData?: User | null;
  isLoading?: boolean;
}

export const UserForm: React.FC<UserFormProps> = ({ onSubmit, initialData, isLoading }) => {
  const { t } = useTranslation("common");
  const [name, setName] = useState(initialData?.name ?? "");
  const [email, setEmail] = useState(initialData?.email ?? "");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(name, email, initialData?.id);
    if (!initialData) {
      // Clear form only for new user creation
      setName("");
      setEmail("");
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3"
    >
      <Input
        type="text"
        placeholder={t("labels.fullName")}
        value={name}
        onChange={(e) => setName(e.target.value)}
        required
        disabled={isLoading}
      />
      <Input
        type="email"
        placeholder={t("labels.email")}
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
        disabled={isLoading}
      />
      <Button
        type="submit"
        disabled={isLoading}
      >
        {initialData ? t("buttons.save") : t("buttons.save")}
      </Button>
    </form>
  );
};
