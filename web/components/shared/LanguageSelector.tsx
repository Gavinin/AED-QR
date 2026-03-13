"use client";

import { useTranslation } from "react-i18next";
import { 
  Modal, 
  ModalContent, 
  ModalHeader, 
  ModalBody, 
  Button, 
  useDisclosure,
} from "@heroui/react";
import { FaGlobe } from "react-icons/fa";

interface Language {
  code: string;
  name: string;
  nativeName: string;
}

const LANGUAGES: Language[] = [
  { code: "en", name: "English", nativeName: "English" },
  { code: "zh", name: "Chinese", nativeName: "中文" },
  { code: "ja", name: "Japanese", nativeName: "日本語" },
  { code: "ko", name: "Korean", nativeName: "한국어" },
  { code: "es", name: "Spanish", nativeName: "Español" },
  { code: "ar", name: "Arabic", nativeName: "العربية" },
  { code: "fr", name: "French", nativeName: "Français" },
];

interface LanguageSelectorProps {
  variant?: "ghost" | "light" | "flat" | "solid" | "bordered" | "faded" | "shadow";
  className?: string;
  showLabel?: boolean;
  customTrigger?: (onOpen: () => void) => React.ReactNode;
}

export default function LanguageSelector({ variant = "light", className, showLabel = false, customTrigger }: LanguageSelectorProps) {
  const { t, i18n } = useTranslation();
  const { isOpen, onOpen, onOpenChange } = useDisclosure();

  const handleLanguageSelect = (code: string) => {
    i18n.changeLanguage(code);
    // onOpenChange(false); // Close modal is handled by Listbox selection if we want, or manually
  };

  const currentLang = LANGUAGES.find(l => l.code === i18n.language) || LANGUAGES[0];

  return (
    <>
      {customTrigger ? (
        customTrigger(onOpen)
      ) : (
        <Button 
          isIconOnly={!showLabel}
          variant={variant} 
          onPress={onOpen} 
          aria-label="Switch Language" 
          className={className}
        >
          <FaGlobe size={20} />
          {showLabel && <span className="ml-2">{currentLang.nativeName}</span>}
        </Button>
      )}

      <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="sm">
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader className="flex flex-col gap-1">
                {t("common.select_language", "Select Language")}
              </ModalHeader>
              <ModalBody className="pb-6">
                <div className="flex flex-col gap-1 w-full">
                  {LANGUAGES.map((lang) => (
                    <Button
                      key={lang.code}
                      variant={i18n.language === lang.code ? "flat" : "light"}
                      color={i18n.language === lang.code ? "primary" : "default"}
                      onPress={() => {
                        handleLanguageSelect(lang.code);
                        onClose();
                      }}
                      className="justify-between h-12 w-full"
                    >
                      <div className="flex justify-between items-center w-full px-2">
                        <span className="text-medium font-medium">{lang.nativeName}</span>
                        <span className="text-small text-default-400">{lang.name}</span>
                      </div>
                    </Button>
                  ))}
                </div>
              </ModalBody>
            </>
          )}
        </ModalContent>
      </Modal>
    </>
  );
}
