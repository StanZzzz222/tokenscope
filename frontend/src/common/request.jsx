import axios from 'axios';
import { addToast } from '@heroui/react';
import { IoInformationCircle } from 'react-icons/io5';

export const createRequestInstance = (baseURL, i18n) => {
  const instance = axios.create({
    baseURL: baseURL,
    timeout: 30000,
    headers: { 'Content-Type': 'application/json' },
  });

  instance.interceptors.response.use(
    (response) => response.data,
    (error) => {
      let message = i18n.get("error_request");
      if (error.response) {
        const status = error.response.status;
        switch (status) {
          case 404:
            message = i18n.get("error_not_found");
            break;
          case 500:
            message = i18n.get("error_internal");
            break;
          default:
            message = `${i18n.get("error_request_failed")} ${status}`;
        }
      } else if (error.code === 'ECONNABORTED') {
        message = i18n.get("error_request_timeout");
      } else {
        message = error.message || 'Unknown';
      }

      addToast({
        title: i18n.get("error_title"),
        description: message,
        color: "danger",
        icon: <IoInformationCircle />,
        promise: new Promise((resolve) => setTimeout(resolve, 3000)),
        variant: "solid",
      });

      return Promise.reject(error);
    }
  );

  return instance;
};
