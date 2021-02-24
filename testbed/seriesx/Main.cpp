//
// Main.cpp
//

#include "pch.h"
#include "Game.h"
#include "next.h"

#include <appnotify.h>

using namespace DirectX;

const char* customer_public_key = "pX89DxRQBaQuYebn80myIh+12DPY2mXFhoRhiLRpKFoffKSVE0zotg==";

extern const char* next_log_level_str(int level)
{
    if (level == NEXT_LOG_LEVEL_DEBUG)
        return "debug";
    else if (level == NEXT_LOG_LEVEL_INFO)
        return "info";
    else if (level == NEXT_LOG_LEVEL_ERROR)
        return "error";
    else if (level == NEXT_LOG_LEVEL_WARN)
        return "warning";
    else
        return "???";
}

void xbox_printf(int level, const char* format, ...)
{
    va_list args;
    va_start(args, format);
    char buffer[1024];
    vsnprintf(buffer, sizeof(buffer), format, args);
    const char* level_str = next_log_level_str(level);
    char buffer2[1024];
    if (level != NEXT_LOG_LEVEL_NONE)
    {
        snprintf(buffer2, sizeof(buffer2), "%0.2f %s: %s\n", next_time(), level_str, buffer);
    }
    else
    {
        snprintf(buffer2, sizeof(buffer2), "%s\n", buffer);
    }
    OutputDebugStringA(buffer2);
    va_end(args);
}

namespace
{
    std::unique_ptr<Game> g_game;
    HANDLE g_plmSuspendComplete = nullptr;
    HANDLE g_plmSignalResume = nullptr;
};

LRESULT CALLBACK WndProc(HWND, UINT, WPARAM, LPARAM);

// Entry point
int WINAPI wWinMain(_In_ HINSTANCE hInstance, _In_opt_ HINSTANCE, _In_ LPWSTR lpCmdLine, _In_ int nCmdShow)
{
    UNREFERENCED_PARAMETER(lpCmdLine);

    if (!XMVerifyCPUSupport())
    {
#ifdef _DEBUG
        OutputDebugStringA("ERROR: This hardware does not support the required instruction set.\n");
#ifdef __AVX2__
        OutputDebugStringA("This may indicate a Gaming.Xbox.Scarlett.x64 binary is being run on an Xbox One.\n");
#endif
#endif
        return 1;
    }

    if (FAILED(XGameRuntimeInitialize()))
        return 1;

    // Microsoft Game Core on Xbox supports UTF-8 everywhere
    assert(GetACP() == CP_UTF8);

    OutputDebugStringA("\nRunning tests...\n\n");

    next_log_level(NEXT_LOG_LEVEL_NONE);

    next_config_t config;
    next_default_config(&config);
    strncpy_s(config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1);

    next_log_function( xbox_printf );

    next_init(NULL, &config);

    next_test();

    OutputDebugStringA("\nAll tests passed successfully!\n\n");

    g_game = std::make_unique<Game>();

    // Register class and create window
    PAPPSTATE_REGISTRATION hPLM = {};
    {
        // Register class
        WNDCLASSEXA wcex = {};
        wcex.cbSize = sizeof(WNDCLASSEXA);
        wcex.style = CS_HREDRAW | CS_VREDRAW;
        wcex.lpfnWndProc = WndProc;
        wcex.hInstance = hInstance;
        wcex.lpszClassName = u8"gdkWindowClass";
        wcex.hbrBackground = reinterpret_cast<HBRUSH>(COLOR_WINDOW + 1);
        if (!RegisterClassExA(&wcex))
            return 1;

        // Create window
        HWND hwnd = CreateWindowExA(0, u8"gdkWindowClass", u8"gdk", WS_OVERLAPPEDWINDOW,
            CW_USEDEFAULT, CW_USEDEFAULT, 1920, 1080, nullptr, nullptr, hInstance,
            nullptr);
        if (!hwnd)
            return 1;

        ShowWindow(hwnd, nCmdShow);

        SetWindowLongPtr(hwnd, GWLP_USERDATA, reinterpret_cast<LONG_PTR>(g_game.get()));

        g_game->Initialize(hwnd);

        g_plmSuspendComplete = CreateEventEx(nullptr, nullptr, 0, EVENT_MODIFY_STATE | SYNCHRONIZE);
        g_plmSignalResume = CreateEventEx(nullptr, nullptr, 0, EVENT_MODIFY_STATE | SYNCHRONIZE);
        if (!g_plmSuspendComplete || !g_plmSignalResume)
            return 1;

        if (RegisterAppStateChangeNotification([](BOOLEAN quiesced, PVOID context)
        {
            if (quiesced)
            {
                ResetEvent(g_plmSuspendComplete);
                ResetEvent(g_plmSignalResume);

                // To ensure we use the main UI thread to process the notification, we self-post a message
                PostMessage(reinterpret_cast<HWND>(context), WM_USER, 0, 0);

                // To defer suspend, you must wait to exit this callback
                (void)WaitForSingleObject(g_plmSuspendComplete, INFINITE);
            }
            else
            {
                SetEvent(g_plmSignalResume);
            }
        }, hwnd, &hPLM))
            return 1;
    }

    // Main message loop
    MSG msg = {};
    while (WM_QUIT != msg.message)
    {
        if (PeekMessage(&msg, nullptr, 0, 0, PM_REMOVE))
        {
            TranslateMessage(&msg);
            DispatchMessage(&msg);
        }
        else
        {
            g_game->Tick();
        }
    }

    g_game.reset();

    UnregisterAppStateChangeNotification(hPLM);

    CloseHandle(g_plmSuspendComplete);
    CloseHandle(g_plmSignalResume);

    XGameRuntimeUninitialize();

    next_term();

    return static_cast<int>(msg.wParam);
}

// Windows procedure
LRESULT CALLBACK WndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam)
{
    auto game = reinterpret_cast<Game*>(GetWindowLongPtr(hWnd, GWLP_USERDATA));

    switch (message)
    {
    case WM_CREATE:
        break;

    case WM_ACTIVATEAPP:
        break;

    case WM_USER:
        if (game)
        {
            game->OnSuspending();

            // Complete deferral
            SetEvent(g_plmSuspendComplete);

            (void)WaitForSingleObject(g_plmSignalResume, INFINITE);

            game->OnResuming();
        }
        break;
    }

    return DefWindowProc(hWnd, message, wParam, lParam);
}

// Exit helper
void ExitGame() noexcept
{
    PostQuitMessage(0);
}
