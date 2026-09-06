#pragma once

#include <string>
#include <windows.h>

class HWIDExtractor {
public:
    static std::string getHardwareUUID();
    
private:
    static std::string getSMBIOSUUID();
    static std::string getCPUId();
    static std::string getPrimaryMACAddress();
    static std::string generateSHA256(const std::string& input);
};