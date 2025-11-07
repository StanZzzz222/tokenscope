import { useState } from "react";
import {
  Accordion,
  AccordionItem,
  Button,
  Card,
  CardBody,
  CardHeader,
  Chip,
  CircularProgress,
  Image,
  Input,
  Modal,
  ModalBody,
  ModalContent,
  ModalHeader,
  Pagination,
  Skeleton,
  Spacer,
  Tab,
  Tabs,
  Textarea,
} from "@heroui/react";
import { IoAddOutline, IoArrowBackOutline, IoCopy, IoInformationCircle, IoRemoveOutline, IoSearch } from "react-icons/io5";
import { useI18n } from "../../common/i18n";
import { useToast } from "../../common/toast";
import { createAPIInstance } from "../../common/api";
import { formatEther, formatUnits, isAddress } from "ethers";
import BlockiesSvg from "blockies-react-svg";
import NoData from "../../components/NoData";
import { baseURL } from "../../common/config"
import { useEffect } from "react";

const Index = () => {
  const i18n = useI18n();
  const apiInstance = createAPIInstance(i18n);
  const { toastError } = useToast();
  // State
  const [initAddress, setInitAddress] = useState("");
  const [address, setAddress] = useState("");
  const [asset, setAsset] = useState(null);
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(false);
  const [erc20TokenAssets, setErc20TokenAssets] = useState(new Map())
  const [erc721TokenAssets, setErc721TokenAssets] = useState(new Map())
  const [nftImages, setNftImages] = useState(new Map())
  const [page, setPage] = useState(1)

  useEffect(() => {
    const ret = localStorage.getItem("address")
    if (ret) setInitAddress(ret)
  }, [])

  useEffect(() => {
    if (initAddress.length > 0) {
      handleSubmit(initAddress)
      setAddress(initAddress)
    }
  }, [initAddress])

  const getTokenValue = async (address, tokenAddress, decimals) => {
     const ret = await apiInstance.get(`/asset/erc20_token_assets/${address}/${tokenAddress}`);
     return Number.parseFloat(formatUnits(BigInt(ret?.data?.value), decimals)).toFixed(6)
  }

  const getNftTokens = async (address, tokenAddress) => {
     const ret = await apiInstance.get(`/asset/erc721_token_assets/${address}/${tokenAddress}`);
     return ret?.data
  }

  const handleSubmit = async (address) => {
    setLoading(true);
    const isVerify = isAddress(address);
    if (address.length < 42 || !isVerify) {
      toastError(i18n.get("address_error"));
      return;
    }
    const ret = await apiInstance.get(`/asset/${address}`);
    setAsset(ret.data);
    setLoading(false);
    localStorage.setItem("address", address)
    for (const token of ret.data.erc20_tokens) {
      const value = await getTokenValue(address, token.address, token.decimals);
      setErc20TokenAssets((prev) => {
        const newMap = new Map(prev);
        newMap.set(token.address, value);
        return newMap;
      });
    }
  }

  const handleChangeNFT = async (keys) => {
    const arr = Array.from(keys)
    const key = arr[0]
    if (key && !erc721TokenAssets.has(key)) {
      const tokens = await getNftTokens(address, key)
      setErc721TokenAssets((prev) => {
        const newMap = new Map(prev);
        newMap.set(key, tokens);
        return newMap;
      })
      for (const token of tokens) {
        getNftImage(token).then((image) => {
          setNftImages((prev) => {
            const newMap = new Map(prev);
            const current = newMap.get(key) ?? [];
            newMap.set(key, [...current, image]);
            return newMap;
          });
        });
      }
    }
  }

  const handleClose = () => {
    setPage(1)
    setAsset(null)
    setDetail(null)
    setAddress("")
    setNftImages(new Map())
    setErc20TokenAssets(new Map())
    setErc721TokenAssets(new Map())
    localStorage.removeItem("address")
  }

  const handleChangeAddress = (el) => {
    const value = el.target.value;
    setAddress(value);
  }

  const handleDetail = (item) => {
    setDetail(item)
  }

  const getTag = (item) => {
    let tagColor = "success"
    if (item.from.toLowerCase() === address.toLowerCase()) tagColor = "danger"
    return <Chip size="lg" color={tagColor} radius="sm">{getTagType(item) === "sender" ? i18n.get("sender") : i18n.get("reciever")}</Chip>
  }

  const getTagType = (item) => {
    let tag = "reciever"
    if (item.from.toLowerCase() === address.toLowerCase()) tag = "sender"
    return tag
  }

  const getNftImage = async (item) => {
    if (!item) return null
    if (item.special_url) return item.metadata_url ?? null
    let { data } = await apiInstance.get(`/asset/metadata/${item?.token?.address}/${item?.token_id}`)
    if (data?.image.indexOf("ipfs://") !== -1) data.image = data?.image.replaceAll("ipfs://", `https://gateway.pinata.cloud/ipfs/`)
    return data?.image ?? null
  }


  const formatTimestamp = (ts) => {
    if (ts.toString().length <= 10) ts = ts * 1000;
    const date = new Date(ts);
    const Y = date.getFullYear();
    const M = String(date.getMonth() + 1).padStart(2, '0');
    const D = String(date.getDate()).padStart(2, '0');
    const h = String(date.getHours()).padStart(2, '0');
    const m = String(date.getMinutes()).padStart(2, '0');
    const s = String(date.getSeconds()).padStart(2, '0');
    return `${Y}-${M}-${D} ${h}:${m}:${s}`;
  }

  return (
    <>
      <div className="w-screen h-screen overflow-auto bg-default-100 min-h-screen">
        <div className="w-full h-full flex items-center justify-center">
          <div className="w-[50%] h-full flex flex-col items-center justify-center space-y-10">
            {asset === null ? (
              <>
                <div className="text-[2.4vh]">{i18n.get("index_title")}</div>
                <div className="w-full flex flex-row items-center justify-center space-x-4">
                  <Input
                    disabled={loading}
                    onChange={handleChangeAddress}
                    startContent={<IoSearch fontSize={18} />}
                    size="lg"
                    variant="faded"
                    value={address ?? initAddress}
                    placeholder={i18n.get("search_placeholder")}
                    type="text"
                  />
                  <Button
                    isLoading={loading}
                    radius="sm"
                    onPress={() => { handleSubmit(address) }}
                    size="lg"
                    color="primary"
                  >
                    {i18n.get("search")}
                  </Button>
                </div>
              </>
            ) : (
              <>
                <Card className="w-screen h-screen bg-default-100 items-center justify-center">
                  <CardHeader className="w-[95%] mt-8 flex items-center justify-center space-x-4">
                    <BlockiesSvg
                      address={address}
                      className="w-[5vh] h-[5vh] rounded-[1vh]"
                    />
                    <div className="w-full flex flex-col p-2 justify-between">
                      <p className="flex flex-row items-center justify-start text-[1.65vh] text-default-600">
                        {asset.address}
                        <Spacer />
                        <Button size="sm" isIconOnly>
                          <IoCopy fontSize={16} />
                        </Button>
                      </p>
                      <p className="text-[1.85vh] font-bold text-default-800">
                        {Number.parseFloat(
                          formatEther(BigInt(asset.balance))
                        ).toFixed(8)}&nbsp;
                        ETH
                      </p>
                    </div>
                    <Button onPress={handleClose} size="lg" radius="sm" variant="flat" className="text-[1.55vh] bg-default-200 text-default-600">
                      <IoArrowBackOutline fontSize={30} />
                      {i18n.get("back")}
                    </Button>
                  </CardHeader>
                  <CardBody className="flex items-center">
                    <Tabs
                      variant="light"
                      size="lg"
                      radius="sm"
                      defaultSelectedKey={`erc20`}
                      className="w-[95%]"
                    >
                      <Tab className="w-[95%]" key={`erc20`} title={i18n.get("erc20_tokens")}>
                        <div className="w-full h-[75vh] space-y-4 overflow-auto">
                          {asset.erc20_tokens.length > 0 ? (
                            <>
                              {asset.erc20_tokens.map((item, key) => {
                                return (
                                  <Card
                                    isPressable
                                    key={`${key}`}
                                    className="w-full h-[8vh] flex items-center text-[1.65vh] p-4 bg-default-200 rounded-[0.5vh]"
                                  >
                                    <div className="w-full flex flex-row items-center justify-between">
                                      <div className="flex flex-row items-center space-x-4">
                                        <Image
                                          className="w-[5vh] h-[5vh] bg-default-400"
                                          radius="full"
                                          src={`${baseURL}/asset/icon/${item.address}`}
                                        />
                                        <div className="flex flex-col items-start justify-start">
                                          <div className="text-[1.65vh] text-default-800">{item.symbol}</div>
                                          <div className="text-[1.35vh] text-default-600">{item.name}</div>
                                        </div>
                                      </div>
                                      <div className="w-[15vh] max-h-[5vh] min-h-[5vh] rounded-sm">
                                        <Skeleton
                                          isLoaded={erc20TokenAssets.get(item.address) ? true : false}
                                          className="rounded-lg"
                                        >
                                          <div className="w-full min-h-[5vh] text-[2vh] text-[#ececec] flex flex-col justify-center items-end">
                                            {erc20TokenAssets.get(item.address)}
                                          </div>
                                        </Skeleton>
                                      </div>
                                    </div>
                                  </Card>
                                );
                              })}
                            </>
                          ) : (
                            <NoData />
                          )}
                        </div>
                      </Tab>
                      <Tab className="w-[95%]" key={`erc721`} title={i18n.get("erc721_tokens")}>
                        <div className="w-full h-[75vh] space-y-4 overflow-auto">
                          {asset.erc721_tokens.length > 0 ? (
                            <Accordion className="w-full overflow-auto" onSelectionChange={handleChangeNFT} variant="bordered">
                              {asset.erc721_tokens.map((item, key) => {
                                return (
                                  <AccordionItem
                                    className="w-full max-h-[65vh] p-2"
                                    key={`${item.address}`}
                                    title={
                                      <div className="text-[1.85vh] font-bold">
                                        {item.name}
                                      </div>
                                    }
                                  >
                                    {erc721TokenAssets.size > 0 ? (
                                      <>
                                        <div key={`${key}`} className="w-full h-full">
                                          {erc721TokenAssets.get(item.address) && erc721TokenAssets.get(item.address).length > 0 ? (
                                            <>
                                              <div className="gap-4 grid grid-cols-8">
                                                {erc721TokenAssets.get(item.address).map((item, key) => {
                                                  return (
                                                    <Card className="bg-default-200" shadow="none" key={`${key}`} isPressable onPress={() => console.log("item pressed")}>
                                                      {nftImages.get(item?.token?.address) ? (
                                                        <Image
                                                          alt={item.title}
                                                          radius="none"
                                                          className="w-full h-[20vh] object-cover"
                                                          src={nftImages.get(item?.token?.address)?.[key]}
                                                          width="100%"
                                                        />
                                                      ) : (
                                                        <Skeleton>
                                                          <div className="w-full h-[20vh] bg-default-300" />
                                                        </Skeleton>
                                                      )}
                                                      <div className="p-4 bg-default-200">
                                                        <b className="text-[1.4vh]">{item?.token?.name}</b>
                                                        <Spacer />
                                                        <p className="text-default-500">
                                                          #{item?.token_id?.toString().length > 16
                                                            ? item.token_id.toString().slice(0, 16) + "…"
                                                            : item.token_id.toString()}
                                                        </p>
                                                      </div>
                                                    </Card>
                                                  )
                                                })}
                                              </div>
                                            </>
                                          ) : <NoData />}
                                        </div>
                                      </>
                                    ) : (
                                      <div className="w-full min-h-[10vh] flex flex-col items-center justify-center space-y-2">
                                        <CircularProgress size="lg" color="default-400" />
                                        <div className="text-[1.65vh] text-default-600">{i18n.get("loading")}</div>
                                        <div className="text-[1.35vh] text-danger-600">{i18n.get("nft_loading_note")}</div>
                                      </div>
                                    )}
                                  </AccordionItem>
                                );
                              })}
                            </Accordion>
                          ) : (
                            <NoData />
                          )}
                        </div>
                      </Tab>
                      <Tab className="w-[95%]" key={`txs`} title={i18n.get("txs")}>
                        <Modal
                          hideCloseButton
                          classNames={{
                            backdrop: "bg-[#1d1d1d]/60 backdrop-opacity-40",
                            base: "p-6 bg-default-200 dark:bg-default-200",
                          }} size="3xl" 
                          onClose={() => setDetail(null)} 
                          isOpen={detail ? true : false}
                        >
                          <ModalContent>
                            <ModalHeader className="flex flex-col gap-1 text-[2vh]">{i18n.get("tx_detail_title")}</ModalHeader>
                            <ModalBody>
                              <div className="flex flex-row justify-between">
                                <div className="text-[1.45vh]">{i18n.get("sender")}</div>
                                <div className="text-[1.25vh] text-default-500 flex items-center">
                                  {detail?.from}&nbsp;
                                  <Button size="sm" isIconOnly>
                                    <IoCopy fontSize={16} />
                                  </Button>
                                </div>
                              </div>
                              <div className="flex flex-row justify-between">
                                <div className="text-[1.45vh]">{i18n.get("reciever")}</div>
                                <div className="text-[1.25vh] text-default-500 flex items-center">
                                  {detail?.to}&nbsp;
                                  <Button size="sm" isIconOnly>
                                    <IoCopy fontSize={16} />
                                  </Button>
                                </div>
                              </div>
                              <div className="flex flex-row justify-between">
                                <div className="text-[1.45vh]">{i18n.get("amount")}</div>
                                <div className={`text-[1.65vh] ${detail ? getTagType(detail) !== "reciever" ? "text-danger-500" : "text-success-500" : null} flex flex-row space-x-4 items-center`}>
                                  {detail ? (
                                    <>
                                      {getTagType(detail) !== "reciever" ? <IoRemoveOutline /> : <IoAddOutline />}&nbsp;
                                      {detail.value !== "0" ? Number.parseFloat(formatEther(BigInt(detail?.value))).toFixed(8) : 0}&nbsp;ETH
                                    </>
                                  ): null}
                                </div>
                              </div>
                              <div className="flex flex-row justify-between">
                                <div className="text-[1.45vh]">{i18n.get("datetime")}</div>
                                <div className={`text-[1.45vh] text-default-800 flex flex-row space-x-4 items-center`}>
                                  {detail ? formatTimestamp(detail?.timestamp) : null}
                                </div>
                              </div>
                              <div className="flex flex-col space-y-2 mt-6">
                                <div className="text-[1.45vh] flex items-center">
                                  {i18n.get("source_data")}&nbsp;
                                  <Button size="sm" isIconOnly>
                                    <IoCopy fontSize={16} />
                                  </Button>
                                </div>
                                <Textarea size="lg" isDisabled value={detail?.data}></Textarea>
                              </div>
                            </ModalBody>
                          </ModalContent>
                        </Modal>
                        <div className="w-full h-[75vh] space-y-4 overflow-auto">
                          {asset.txs.length > 0 ? (
                            <>
                              {asset.txs.slice((page-1)*10,page * 10).map((item, key) => {
                                return (
                                  <div key={`${key}`} className="flex flex-row items-center space-x-4">
                                    {getTag(item)}
                                    <div className="w-full flex flex-row items-center justify-between">
                                      <div className="flex flex-col">
                                        <div className="text-[1.65vh] text-default-800">{item.from}</div>
                                        <div className="text-[1.35vh] text-default-600">{item.to}</div>
                                      </div>
                                      <div className="flex flex-row items-center space-x-4">
                                        <div className="min-w-[15vh] flex flex-col justify-center items-end">
                                          <div className={`text-[1.65vh] ${getTagType(item) !== "reciever" ? "text-danger-500" : "text-success-500"} flex flex-row space-x-4 items-center`}>
                                            {getTagType(item) !== "reciever" ? <IoRemoveOutline /> : <IoAddOutline />}&nbsp;
                                            {item.value !== "0" ? Number.parseFloat(formatEther(BigInt(item.value))).toFixed(8) : 0}&nbsp;ETH
                                          </div>
                                          <div className="text-[1.35vh] text-default-500">{formatTimestamp(item.timestamp)}</div>
                                        </div>
                                        <Button onPress={() => handleDetail(item)} radius="sm" variant="flat" color="success">
                                          <IoInformationCircle fontSize={16} />
                                          {i18n.get("detail")}
                                        </Button>
                                      </div>
                                    </div>
                                  </div>
                                );
                              })}
                              <div className="w-full flex justify-between mt-[2%]">
                                <div className="text-[1.55vh]">{asset.txs.length} {i18n.get("tx_count")}</div>
                                <Pagination onChange={setPage} showShadow showControls radius="sm" size="lg" initialPage={page} total={Math.ceil(asset.txs.length / 12)} />
                              </div>
                            </>
                          ) : (
                            <NoData />
                          )}
                        </div>
                      </Tab>
                    </Tabs>
                  </CardBody>
                </Card>
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
};

export default Index;
