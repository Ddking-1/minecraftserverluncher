#include "HWIDExtractor.h"
#include <windows.h>
#include <wmiutils.h>
#include <comdef.h>
#include <intrin.h>
#include <iphlpapi.h>
#include <sstream>
#include <iomanip>
#include <algorithm>
#include <openssl/sha.h>

#pragma comment(lib, "wbemuuid.lib")
#pragma comment(lib, "iphlpapi.lib")

std::string HWIDExtractor::getHardwareUUID() {
    std::string uuid = getSMBIOSUUID();
    if (uuid.empty()) {
        // Fallback to CPU ID
        uuid = getCPUId();
    }
    
    std::string macAddress = getPrimaryMACAddress();
    if (!macAddress.empty()) {
        uuid += macAddress;
    }
    
    return generateSHA256(uuid);
}

std::string HWIDExtractor::getSMBIOSUUID() {
    try {
        // Use WMI to retrieve SMBIOS UUID
        IWbemLocator* pLoc = nullptr;
        IWbemServices* pSvc = nullptr;
        
        HRESULT hr = CoInitializeEx(0, COINIT_MULTITHREADED);
        if (FAILED(hr)) return "";
        
        hr = CoCreateInstance(
            CLSID_WbemLocator,
            0,
            CLSCTX_INPROC_SERVER,
            IID_IWbemLocator,
            (void**)&pLoc
        );
        if (FAILED(hr) || !pLoc) {
            CoUninitialize();
            return "";
        }
        
        hr = pLoc->ConnectServer(
            _bstr_t(L"ROOT\\CIMV2"),
            NULL,
            NULL,
            0,
            0,
            0,
            0,
            &pSvc
        );
        if (FAILED(hr) || !pSvc) {
            pLoc->Release();
            CoUninitialize();
            return "";
        }
        
        IEnumWbemClassObj* pEnumerator = nullptr;
        hr = pSvc->ExecQuery(
            bstr_t(L"WQL"),
            bstr_t(L"SELECT UUID FROM Win32_ComputerSystemProduct"),
            WBEM_FLAG_FORWARD_ONLY | WBEM_FLAG_RETURN_IMMEDIATE,
            0,
            &pEnumerator
        );
        
        if (SUCCEEDED(hr) && pEnumerator) {
            IWbemClassObject* pclsObj = nullptr;
            ULONG uReturn = 0;
            
            while (pEnumerator) {
                hr = pEnumerator->Next(WBEM_INFINITE, 1, &pclsObj, &uReturn);
                if (FAILED(hr) || !uReturn) break;
                
                VARIANT vtProp;
                hr = pclsObj->Get(L"UUID", 0, &vtProp, 0, 0);
                if (SUCCEEDED(hr) && vtProp.vt == VT_BSTR) {
                    std::wstring uuid(vtProp.bstrVal);
                    std::string result(uuid.begin(), uuid.end());
                    
                    VariantClear(&vtProp);
                    pclsObj->Release();
                    pEnumerator->Release();
                    pSvc->Release();
                    pLoc->Release();
                    CoUninitialize();
                    
                    return result;
                }
                
                if (pclsObj) {
                    pclsObj->Release();
                }
            }
            
            pEnumerator->Release();
        }
        
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
    } catch (...) {
        // Handle any exceptions
    }
    
    return "";
}

std::string HWIDExtractor::getCPUId() {
    int cpuInfo[4] = { 0 };
    __cpuid(cpuInfo, 0);
    
    std::ostringstream oss;
    for (int i = 0; i < 4; ++i) {
        oss << std::hex << std::setfill('0') << std::setw(8) << cpuInfo[i];
    }
    
    return oss.str();
}

std::string HWIDExtractor::getPrimaryMACAddress() {
    ULONG bufferLength = 0;
    IP_ADAPTER_INFO* pAdapterInfo = nullptr;
    
    // Get the required buffer size
    if (GetAdaptersInfo(NULL, &bufferLength) == ERROR_BUFFER_OVERFLOW) {
        pAdapterInfo = (IP_ADAPTER_INFO*)malloc(bufferLength);
        if (!pAdapterInfo) return "";
        
        if (GetAdaptersInfo(pAdapterInfo, &bufferLength) == NO_ERROR) {
            // Return the MAC address of the first adapter
            std::ostringstream oss;
            for (int i = 0; i < 6; ++i) {
                oss << std::hex << std::setfill('0') << std::setw(2) 
                    << static_cast<int>(pAdapterInfo->PhysicalAddress[i]);
                if (i < 5) oss << "-";
            }
            
            free(pAdapterInfo);
            return oss.str();
        }
        
        free(pAdapterInfo);
    }
    
    return "";
}

std::string HWIDExtractor::generateSHA256(const std::string& input) {
    unsigned char hash[32];
    SHA256_CTX sha256;
    
    if (SHA256_Init(&sha256)) {
        if (SHA256_Update(&sha256, input.c_str(), input.length())) {
            if (SHA256_Final(hash, &sha256)) {
                std::ostringstream oss;
                for (int i = 0; i < 32; ++i) {
                    oss << std::hex << std::setfill('0') << std::setw(2)
                        << static_cast<int>(hash[i]);
                }
                return oss.str();
            }
        }
    }
    
    return "";
}