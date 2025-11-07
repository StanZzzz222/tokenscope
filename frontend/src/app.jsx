import { HeroUIProvider, ToastProvider } from "@heroui/react";
import Index from "./pages/Index";
import Bottom from "./pages/Bottom"
import { I18nProvider } from "./common/i18n";

const App = () => {
  return (
    <>
      <I18nProvider>
        <HeroUIProvider>
          <ToastProvider toastOffset={30} placement="top-center" maxVisibleToasts={1} />
          <Index />
          <Bottom />
        </HeroUIProvider>
      </I18nProvider>
    </>
  );
}

export default App;
