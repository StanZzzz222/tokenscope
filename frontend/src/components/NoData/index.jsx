import { IoCubeOutline } from "react-icons/io5";
import { useI18n } from "../../common/i18n";

const Index = () => {
  const i18n = useI18n();

  return (
    <>
      <div className="w-full h-full flex items-center justify-center mb-[4vh]">
        <div className="flex flex-col items-center justify-center space-y-4">
          <IoCubeOutline size={65} />
          <div className="text-[2.65vh]">{i18n.get("no_data")}</div>
        </div>
      </div>
    </>
  );
};

export default Index;
