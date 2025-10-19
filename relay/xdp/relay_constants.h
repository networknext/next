/*
    Network Next XDP Relay
*/

#ifndef RELAY_CONSTANTS_H
#define RELAY_CONSTANTS_H

#define MAX_RELAYS                                                                            1024

#define MAX_SESSIONS                                                                        100000

#define RELAY_HEADER_BYTES                                                                      25

#define RELAY_MTU                                                                             1200

#define RELAY_MAX_PACKET_BYTES                                                                1384

#define RELAY_ADDRESS_NONE                                                                       0
#define RELAY_ADDRESS_IPV4                                                                       1
#define RELAY_ADDRESS_IPV6                                                                       2

#define RELAY_OK                                                                                 0
#define RELAY_ERROR                                                                             -1

#define RELAY_MAX_UPDATE_ATTEMPTS                                                               30

#define RELAY_RESPONSE_MAX_BYTES                                              ( 10 * 1024 * 1024 )

#define RELAY_HASH_SIZE                                                         ( MAX_RELAYS * 3 )

#define RELAY_PING_STATS_WINDOW                                                               10.0
#define RELAY_PING_HISTORY_SIZE                                                                 64
#define RELAY_PING_SAFETY                                                                      1.0
#define RELAY_PING_TIME                                                                        0.1

#define RELAY_PING_TOKEN_BYTES                                                                  32
#define RELAY_PING_KEY_BYTES                                                                    32
#define RELAY_SESSION_PRIVATE_KEY_BYTES                                                         32
#define RELAY_ROUTE_TOKEN_BYTES                                                                 71
#define RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES                                                      111
#define RELAY_CONTINUE_TOKEN_BYTES                                                              17
#define RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES                                                    57
#define RELAY_PUBLIC_KEY_BYTES                                                                  32
#define RELAY_PRIVATE_KEY_BYTES                                                                 32
#define RELAY_SECRET_KEY_BYTES                                                                  32
#define RELAY_BACKEND_PUBLIC_KEY_BYTES                                                          32

#define RELAY_ROUTE_REQUEST_PACKET                                                               1
#define RELAY_ROUTE_RESPONSE_PACKET                                                              2
#define RELAY_CLIENT_TO_SERVER_PACKET                                                            3
#define RELAY_SERVER_TO_CLIENT_PACKET                                                            4
#define RELAY_SESSION_PING_PACKET                                                                5
#define RELAY_SESSION_PONG_PACKET                                                                6
#define RELAY_CONTINUE_REQUEST_PACKET                                                            7
#define RELAY_CONTINUE_RESPONSE_PACKET                                                           8
#define RELAY_CLIENT_PING_PACKET                                                                 9
#define RELAY_CLIENT_PONG_PACKET                                                                10
#define RELAY_PING_PACKET                                                                       11
#define RELAY_PONG_PACKET                                                                       12
#define RELAY_SERVER_PING_PACKET                                                                13
#define RELAY_SERVER_PONG_PACKET                                                                14

#define RELAY_COUNTER_PACKETS_SENT                                                               0
#define RELAY_COUNTER_PACKETS_RECEIVED                                                           1
#define RELAY_COUNTER_BYTES_SENT                                                                 2
#define RELAY_COUNTER_BYTES_RECEIVED                                                             3
#define RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET                                         4
#define RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET                                      5
#define RELAY_COUNTER_SESSION_CREATED                                                            6
#define RELAY_COUNTER_SESSION_CONTINUED                                                          7
#define RELAY_COUNTER_SESSION_DESTROYED                                                          8

#define RELAY_COUNTER_RELAY_PING_PACKET_SENT                                                    10
#define RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED                                                11
#define RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY                                          12
#define RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED                                                 13
#define RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE                                              14
#define RELAY_COUNTER_RELAY_PING_PACKET_UNKNOWN_RELAY                                           15

#define RELAY_COUNTER_RELAY_PONG_PACKET_SENT                                                    15
#define RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED                                                16
#define RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE                                              17
#define RELAY_COUNTER_RELAY_PONG_PACKET_UNKNOWN_RELAY                                           18

#define RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED                                               20
#define RELAY_COUNTER_CLIENT_PING_PACKET_WRONG_SIZE                                             21
#define RELAY_COUNTER_CLIENT_PING_PACKET_RESPONDED_WITH_PONG                                    22
#define RELAY_COUNTER_CLIENT_PING_PACKET_DID_NOT_VERIFY                                         23
#define RELAY_COUNTER_CLIENT_PING_PACKET_EXPIRED                                                24

#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED                                             30
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE                                           31
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_DECRYPT_ROUTE_TOKEN                        32
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED                                        33
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP                                  34

#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED                                            40
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE                                          41
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION                              42
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_SESSION_EXPIRED                                     43
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED                                    44
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY                               45
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP                             46

#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED                                          50
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE                                        51
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_DECRYPT_CONTINUE_TOKEN                  52
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED                                     53
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_FIND_SESSION                            54
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_SESSION_EXPIRED                                   55
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP                               56

#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED                                         60
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE                                       61
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED                                 62
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION                           63
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED                                  64
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY                            65
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP                          66

#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED                                          70
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL                                         71
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG                                           72
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION                            73
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_SESSION_EXPIRED                                   74
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED                                  75
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_HEADER_DID_NOT_VERIFY                             76
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP                               77

#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED                                          80
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL                                         81
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG                                           82
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION                            83
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED                                   84
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED                                  85
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_HEADER_DID_NOT_VERIFY                             86
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP                           87

#define RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED                                              90
#define RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE                                            91
#define RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_FIND_SESSION                                92
#define RELAY_COUNTER_SESSION_PING_PACKET_SESSION_EXPIRED                                       93
#define RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED                                      94
#define RELAY_COUNTER_SESSION_PING_PACKET_HEADER_DID_NOT_VERIFY                                 95
#define RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP                                   96

#define RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED                                             100
#define RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE                                           101
#define RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_FIND_SESSION                               102
#define RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_EXPIRED                                      103
#define RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED                                     104
#define RELAY_COUNTER_SESSION_PONG_PACKET_HEADER_DID_NOT_VERIFY                                105
#define RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP                              106

#define RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED                                              110
#define RELAY_COUNTER_SERVER_PING_PACKET_WRONG_SIZE                                            111
#define RELAY_COUNTER_SERVER_PING_PACKET_RESPONDED_WITH_PONG                                   112
#define RELAY_COUNTER_SERVER_PING_PACKET_DID_NOT_VERIFY                                        113
#define RELAY_COUNTER_SERVER_PING_PACKET_EXPIRED                                               114

#define RELAY_COUNTER_PACKET_TOO_LARGE                                                         120
#define RELAY_COUNTER_PACKET_TOO_SMALL                                                         121
#define RELAY_COUNTER_DROP_FRAGMENT                                                            122
#define RELAY_COUNTER_DROP_LARGE_IP_HEADER                                                     123
#define RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST                                                124
#define RELAY_COUNTER_DROPPED_PACKETS                                                          125
#define RELAY_COUNTER_DROPPED_BYTES                                                            126
#define RELAY_COUNTER_NOT_IN_WHITELIST                                                         127
#define RELAY_COUNTER_WHITELIST_ENTRY_EXPIRED                                                  128

#define RELAY_COUNTER_SESSIONS                                                                 130
#define RELAY_COUNTER_ENVELOPE_KBPS_UP                                                         131
#define RELAY_COUNTER_ENVELOPE_KBPS_DOWN                                                       132

#define RELAY_NUM_COUNTERS                                                                     150

#define RELAY_VERSION_LENGTH                                                                    32

#define WHITELIST_TIMEOUT                                                                     1000

#define RELAY_ETHERNET_ADDRESS_BYTES                                                             6

#endif // #ifndef RELAY_CONSTANTS_H
