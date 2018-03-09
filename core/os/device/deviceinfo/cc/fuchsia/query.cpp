/*
 * Copyright (C) 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include "../query.h"

#include <cstring>
#include <string.h>

#include <sys/utsname.h>
#include <unistd.h>

namespace query {

void destroyContext() {
    printf("destroyContext NI\n");
}

bool createContext(void* platform_data) {
    printf("createContext NI\n");
    return false;
}

const char* contextError() { return ""; }

bool hasGLorGLES() { return false; }

int numABIs() { return 1; }

void abi(int idx, device::ABI* abi) {
    abi->set_name("x86_64");
    abi->set_os(device::Linux);
    abi->set_architecture(device::X86_64);
    abi->set_allocated_memorylayout(currentMemoryLayout());
}

device::ABI* currentABI() {
    auto out = new device::ABI();
    abi(0, out);
    return out;
}

int cpuNumCores() { return 2; }

const char* gpuName() { return ""; }

const char* gpuVendor() { return ""; }

const char* instanceName() { return "fuchsia-host-name"; }

const char* hardwareName() { return "fuchsia-hardware-name"; }

device::OSKind osKind() { return device::Linux; }

const char* osName() { return "Fuchsia"; }

const char* osBuild() { return ""; }

int osMajor() { return 0; }

int osMinor() { return 0; }

int osPoint() { return 0; }

}  // namespace query
