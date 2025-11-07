import { Progress, Select, SelectItem, Spacer } from "@heroui/react";
import { useEffect, useState } from "react";
import { useI18n } from "../../common/i18n";
import { createAPIInstance } from "../../common/api";

const Locale = () => {
  const i18n = useI18n();
  const apiInstance = createAPIInstance(i18n);
  const [syncInfo, setSyncInfo] = useState(null)

  useEffect(() => {
    init()
    const interval = setInterval(init, 3000)
    return () => clearInterval(interval)
  }, [])

  const init = async () => {
    const ret = await apiInstance.get(`/blockchain/info`);
    setSyncInfo(ret.data)
  }

  const handleChangeLanguage = (el) => {
    const language = el.target.value
    i18n.setLocale(language)
  }

  return (
    <>
      <div className="relative flex items-center justify-center">
        <div className="w-[94%] flex items-center justify-between absolute bottom-5">
          <div className="w-[100vh] h-[2vh] flex items-center justify-center space-x-[8vh]">
            <div className="w-[160vh] flex flex-col items-start justify-start">
              <div className="flex flew-row items-center">
                <div className="w-[0.6vh] h-[0.6vh] bg-green-500 rounded-full"></div>
                <Spacer />
                {i18n.get("network_status")} ({syncInfo ? syncInfo.block_count : 0} {i18n.get("block")})
              </div>
              <div className="flex flex-row items-center">
                <div className="text-default-700">
                  {i18n.get("current_block")}: {syncInfo ? syncInfo.sync_info.current_block_number : 0} | {i18n.get("latest_block")}: {syncInfo ? syncInfo.sync_info.last_block_number : 0}
                </div>
              </div>
            </div>
            <Progress color="success" radius="sm" showValueLabel={true} label={i18n.get("network_sync")} maxValue={100} minValue={0} value={syncInfo ? Math.floor(syncInfo.percent*100) : 0} />
          </div>
          <div className="w-[20vh]">
            <Select
              isRequired
              variant="bordered"
              onChange={handleChangeLanguage}
              defaultSelectedKeys={[i18n.getLocale()]}
            >
              <SelectItem key={"en"}>English</SelectItem>
              <SelectItem key={"zh"}>简体中文</SelectItem>
            </Select>
          </div>
        </div>
      </div>
    </>
  );
};

export default Locale;
