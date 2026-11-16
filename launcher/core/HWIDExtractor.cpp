// SPDX-License-Identifier: GPL-3.0-only
// Project-owned hardware identifier. Windows CNG (bcrypt) SHA-256; no OpenSSL.
#include "HWIDExtractor.h"

#include <windows.h>
#include <bcrypt.h>
#include <iphlpapi.h>
#include <intrin.h>

#include <iomanip>
#include <sstream>
#include <string>
#include <vector>

#ifdef _MSC_VER
#pragma comment(lib, "bcrypt.lib")
#pragma comment(lib, "iphlpapi.lib")
#endif

std::string HWIDExtractor::getSMBIOSUUID()
{
    HKEY hKey = nullptr;
    if (RegOpenKeyExA(HKEY_LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Cryptography", 0,
                      KEY_QUERY_VALUE | KEY_WOW64_64KEY, &hKey) != ERROR_SUCCESS) {
        return {};
    }
    char buffer[256] = {0};
    DWORD size = sizeof(buffer);
    DWORD type = 0;
    const LONG rc = RegQueryValueExA(hKey, "MachineGuid", nullptr, &type,
                                     reinterpret_cast<LPBYTE>(buffer), &size);
    RegCloseKey(hKey);
    if (rc != ERROR_SUCCESS || type != REG_SZ) {
        return {};
    }
    return std::string(buffer);
}

std::string HWIDExtractor::getCPUId()
{
    int cpuInfo[4] = {0};
    __cpuid(cpuInfo, 0);
    std::ostringstream oss;
    oss << std::hex << std::setfill('0')
        << std::setw(8) << cpuInfo[1]
        << std::setw(8) << cpuInfo[3]
        << std::setw(8) << cpuInfo[2];
    return oss.str();
}

std::string HWIDExtractor::getPrimaryMACAddress()
{
    ULONG len = 0;
    if (GetAdaptersInfo(nullptr, &len) != ERROR_BUFFER_OVERFLOW || len == 0) {
        return {};
    }
    std::vector<BYTE> buffer(len);
    PIP_ADAPTER_INFO info = reinterpret_cast<PIP_ADAPTER_INFO>(buffer.data());
    if (GetAdaptersInfo(info, &len) != ERROR_SUCCESS) {
        return {};
    }
    std::ostringstream oss;
    for (UINT i = 0; i < info->AddressLength; ++i) {
        if (i > 0) {
            oss << "-";
        }
        oss << std::hex << std::setfill('0') << std::setw(2)
            << static_cast<int>(info->Address[i]);
    }
    return oss.str();
}

std::string HWIDExtractor::generateSHA256(std::string_view input)
{
    BCRYPT_ALG_HANDLE alg = nullptr;
    if (!BCRYPT_SUCCESS(BCryptOpenAlgorithmProvider(&alg, BCRYPT_SHA256_ALGORITHM,
                                                    MS_PRIMITIVE_PROVIDER, 0))) {
        return {};
    }

    DWORD hashLen = 0;
    DWORD produced = 0;
    if (!BCRYPT_SUCCESS(BCryptGetProperty(alg, BCRYPT_HASH_LENGTH,
                                          reinterpret_cast<PUCHAR>(&hashLen),
                                          sizeof(hashLen), &produced, 0))) {
        BCryptCloseAlgorithmProvider(alg, 0);
        return {};
    }

    BCRYPT_HASH_HANDLE hash = nullptr;
    if (!BCRYPT_SUCCESS(BCryptCreateHash(alg, &hash, nullptr, 0, nullptr, 0, 0))) {
        BCryptCloseAlgorithmProvider(alg, 0);
        return {};
    }

    const PUCHAR data = reinterpret_cast<PUCHAR>(const_cast<char*>(input.data()));
    const ULONG dataLen = static_cast<ULONG>(input.size());
    if (!BCRYPT_SUCCESS(BCryptHashData(hash, data, dataLen, 0))) {
        BCryptDestroyHash(hash);
        BCryptCloseAlgorithmProvider(alg, 0);
        return {};
    }

    std::vector<BYTE> digest(hashLen);
    if (!BCRYPT_SUCCESS(BCryptFinishHash(hash, digest.data(),
                                         static_cast<ULONG>(digest.size()), 0))) {
        BCryptDestroyHash(hash);
        BCryptCloseAlgorithmProvider(alg, 0);
        return {};
    }

    BCryptDestroyHash(hash);
    BCryptCloseAlgorithmProvider(alg, 0);

    std::ostringstream oss;
    for (BYTE b : digest) {
        oss << std::hex << std::setfill('0') << std::setw(2)
            << static_cast<int>(b);
    }
    return oss.str();
}

std::string HWIDExtractor::getHardwareUUID()
{
    const std::string material = getSMBIOSUUID() + getCPUId() + getPrimaryMACAddress();
    return generateSHA256(material);
}
