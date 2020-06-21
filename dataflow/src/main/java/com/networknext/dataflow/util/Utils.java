package com.networknext.dataflow.util;

import java.net.InetAddress;
import java.net.UnknownHostException;
import com.google.protobuf.Timestamp;
import com.networknext.api.address.AddressOuterClass.Address;

public class Utils {
    public static String hexPrint(long val) {
        String b = Long.toUnsignedString(val, 16).toLowerCase();
        while (b.length() < 16) {
            b = "0" + b;
        }
        return b;
    }

    public static String addressToString(Address addr) {
        if (addr.getType() == Address.Type.NONE) {
            return "";
        } else {
            try {
                return InetAddress.getByAddress(addr.getIp().toByteArray()).getHostAddress();
            } catch (UnknownHostException e) {
                return "";
            }
        }
    }

    public static String addressPortToString(Address addr) {
        if (addr.getType() == Address.Type.NONE) {
            return "";
        } else if (addr.getType() == Address.Type.IPV6) {
            try {
                return "[" + InetAddress.getByAddress(addr.getIp().toByteArray()).getHostAddress()
                        + "]:" + Integer.toString(addr.getPort());
            } catch (UnknownHostException e) {
                return "";
            }
        } else {
            try {
                return InetAddress.getByAddress(addr.getIp().toByteArray()).getHostAddress() + ":"
                        + Integer.toString(addr.getPort());
            } catch (UnknownHostException e) {
                return "";
            }
        }
    }

    public static Timestamp toProtobufTimestamp(long timestamp) {
        return Timestamp.newBuilder().setSeconds(timestamp).setNanos(0).build();
    }
}
