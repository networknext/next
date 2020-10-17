//
// Main.cpp
//

#include "pch.h"
#include <ppltasks.h>
#include "next.h"

using namespace concurrency;
using namespace Windows::ApplicationModel;
using namespace Windows::ApplicationModel::Core;
using namespace Windows::ApplicationModel::Activation;
using namespace Windows::UI::Core;
using namespace Windows::Foundation;
using namespace DirectX;

const char * customer_public_key = "pX89DxRQBaQuYebn80myIh+12DPY2mXFhoRhiLRpKFoffKSVE0zotg==";

void packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;
    (void) context;
    (void) packet_data;
    (void) packet_bytes;
    // ...
}

extern const char * next_log_level_str( int level )
{
    if ( level == NEXT_LOG_LEVEL_DEBUG )
        return "debug";
    else if ( level == NEXT_LOG_LEVEL_INFO )
        return "info";
    else if ( level == NEXT_LOG_LEVEL_ERROR )
        return "error";
    else if ( level == NEXT_LOG_LEVEL_WARN )
        return "warning";
    else
        return "???";
}

void xbox_printf( int level, const char * format, ... ) 
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    const char * level_str = next_log_level_str( level );
	char buffer2[1024];
	if (level != NEXT_LOG_LEVEL_NONE)
	{
		snprintf(buffer2, sizeof(buffer2), "%0.2f %s: %s\n", next_time(), level_str, buffer);
	}
	else
	{
		snprintf(buffer2, sizeof(buffer2), "%s\n", buffer);
	}
	OutputDebugStringA( buffer2 );
    va_end( args );
}

ref class ViewProvider sealed : public IFrameworkView
{
private:
    bool exit;
    next_client_t * client;

public:
    ViewProvider() :
        exit(false)
    {
    }

    // IFrameworkView methods
    virtual void Initialize(CoreApplicationView^ applicationView)
    {
        applicationView->Activated +=
            ref new TypedEventHandler<CoreApplicationView^, IActivatedEventArgs^>(this, &ViewProvider::OnActivated);

        CoreApplication::Suspending +=
            ref new EventHandler<SuspendingEventArgs^>(this, &ViewProvider::OnSuspending);

        CoreApplication::Resuming +=
            ref new EventHandler<Platform::Object^>(this, &ViewProvider::OnResuming);

        CoreApplication::DisableKinectGpuReservation = true;

		next_config_t config;
		next_default_config(&config);
		strncpy_s(config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1);

		next_init(NULL, &config);

		next_log_function( xbox_printf );
    }

    virtual void Uninitialize()
    {
		next_term();
	}

    virtual void SetWindow(CoreWindow^ window)
    {
        window->Closed +=
            ref new TypedEventHandler<CoreWindow^, CoreWindowEventArgs^>(this, &ViewProvider::OnWindowClosed);
    }

    virtual void Load(Platform::String^ entryPoint)
    {
    }

	virtual void Run()
	{
		OutputDebugStringA("\nRunning tests...\n\n");

		next_log_level(NEXT_LOG_LEVEL_NONE);

		next_test();

		OutputDebugStringA("\nAll tests passed successfully!\n\n");

		next_log_level(NEXT_LOG_LEVEL_INFO);

		OutputDebugStringA("Starting client...\n\n");

		client = next_client_create( NULL, "0.0.0.0:0", packet_received, NULL );
		if ( !client)
			return;

        next_client_open_session( client, "173.255.241.176:32202" );

        while ( !exit )
        {
            next_client_update( client );

            uint8_t packet_data[32];
            memset( packet_data, 0, sizeof(packet_data) );
            next_client_send_packet( client, packet_data, sizeof( packet_data ) );

            next_sleep( 1.0f / 60.0f );

            CoreWindow::GetForCurrentThread()->Dispatcher->ProcessEvents(CoreProcessEventsOption::ProcessAllIfPresent);
        }

		OutputDebugStringA("\nShutting down...\n\n");

        next_client_destroy( client );
    }

protected:
    // Event handlers
    void OnActivated(CoreApplicationView^ applicationView, IActivatedEventArgs^ args)
    {
        CoreWindow::GetForCurrentThread()->Activate();
    }

    void OnSuspending(Platform::Object^ sender, SuspendingEventArgs^ args)
    {
    }

    void OnResuming(Platform::Object^ sender, Platform::Object^ args)
    {
    }

    void OnWindowClosed(CoreWindow^ sender, CoreWindowEventArgs^ args)
    {
        exit = true;
    }
};

ref class ViewProviderFactory : IFrameworkViewSource
{
public:
    virtual IFrameworkView^ CreateView()
    {
        return ref new ViewProvider();
    }
};


// Entry point
[Platform::MTAThread]
int __cdecl main(Platform::Array<Platform::String^>^ argv)
{
    auto viewProviderFactory = ref new ViewProviderFactory();
    CoreApplication::Run(viewProviderFactory);
    return 0;
}
