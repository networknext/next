//
//  AppDelegate.m
//  sample
//
//  Created by Evan Todd on 6/29/18.
//  Copyright © 2018 Network Next. All rights reserved.
//

#import "AppDelegate.h"
#include "next/next.h"

@interface AppDelegate ()
{
    next_client_t * client;
    double last_report_time;
}
@end

void packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;
    (void) context;
    (void) packet_data;
    (void) packet_bytes;
    // ...
}

static const char * log_level_str( int level )
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
        return "log";
}

static void print_function( int level, const char * format, ...)
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    const char * level_str = log_level_str( level );
    va_end( args );
    NSLog( @"%0.2f %s: %s", next_time(), level_str, buffer );
}

@implementation AppDelegate

- (void)update {
    
    next_client_update( client );
    
    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof(packet_data) );
    next_client_send_packet( client, packet_data, sizeof( packet_data ) );
    
    if ( last_report_time + 1.0f < next_time() )
    {
        const next_client_stats_t * stats = next_client_stats( client );
        if ( stats->next_rtt >= 0.0f && stats->direct_rtt >= 0.0f )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "Direct RTT = %.1fms, Next RTT = %.1fms", stats->direct_rtt, stats->next_rtt );
        }
        else if ( stats->direct_rtt >= 0.0f )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "Direct RTT = %.1fms", stats->direct_rtt );
        }
        else if ( stats->next_rtt >= 0.0f )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "Next RTT = %.1fms", stats->next_rtt );
        }
        last_report_time = next_time();
    }
}


- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
    
    CADisplayLink * display = [CADisplayLink displayLinkWithTarget:self selector:@selector(update)];
    [display addToRunLoop:[NSRunLoop currentRunLoop] forMode:NSRunLoopCommonModes];
    
    next_log_function( print_function );
    
    NSLog( @"\nWelcome to Network Next!\n\n" );
    
    next_init();
    
    next_log_level( NEXT_LOG_LEVEL_INFO );
    
    client = next_client_create( NULL, packet_received );
    if ( !client )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create network next client" );
        exit( 1 );
    }
    
    next_client_open_session( client, "127.0.0.1:50000" );
    
    return YES;
}


- (void)applicationWillResignActive:(UIApplication *)application {
    // Sent when the application is about to move from active to inactive state. This can occur for certain types of temporary interruptions (such as an incoming phone call or SMS message) or when the user quits the application and it begins the transition to the background state.
    // Use this method to pause ongoing tasks, disable timers, and invalidate graphics rendering callbacks. Games should use this method to pause the game.
}


- (void)applicationDidEnterBackground:(UIApplication *)application {
    // Use this method to release shared resources, save user data, invalidate timers, and store enough application state information to restore your application to its current state in case it is terminated later.
    // If your application supports background execution, this method is called instead of applicationWillTerminate: when the user quits.
}


- (void)applicationWillEnterForeground:(UIApplication *)application {
    // Called as part of the transition from the background to the active state; here you can undo many of the changes made on entering the background.
}


- (void)applicationDidBecomeActive:(UIApplication *)application {
    // Restart any tasks that were paused (or not yet started) while the application was inactive. If the application was previously in the background, optionally refresh the user interface.
}


- (void)applicationWillTerminate:(UIApplication *)application {
    // Called when the application is about to terminate. Save data if appropriate. See also applicationDidEnterBackground:.
    
    next_printf( NEXT_LOG_LEVEL_INFO, "stopping client" );
    
    next_client_destroy( client );
    
    next_term();
    
    NSLog( @"\n" );
}


@end
