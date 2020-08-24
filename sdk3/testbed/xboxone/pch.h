//
// pch.h
// Header for standard system include files.
//

#pragma once

// Use the C++ standard templated min/max
#define NOMINMAX

#include <xdk.h>
#include <wrl.h>
#include <d3d12_x.h>
#include <d3dx12_x.h>
#include <DirectXMath.h>
#include <DirectXColors.h>

#include <algorithm>
#include <memory>

#include <pix.h>

namespace DX
{
    inline void ThrowIfFailed(HRESULT hr)
    {
        if (FAILED(hr))
        {
            // Set a breakpoint on this line to catch DirectX API errors
            throw Platform::Exception::CreateException(hr);
        }
    }
}
