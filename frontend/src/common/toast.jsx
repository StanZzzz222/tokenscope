import { addToast } from "@heroui/react";
import { IoInformationCircle } from 'react-icons/io5';
import { useI18n } from "./i18n";

export function useToast() {
  const i18n = useI18n();

  function toastError(message) {
    addToast({
      title: i18n.get("error_title"),
      description: message,
      color: "danger",
      icon: <IoInformationCircle />,
      variant: "solid",
    });
  }
  return { toastError };
}