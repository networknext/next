package com.networknext.dataflow.util.fnv;

import java.math.BigInteger;

public class FNV {
    private static final BigInteger INIT32 = new BigInteger("811c9dc5", 16);
    private static final BigInteger INIT64 = new BigInteger("cbf29ce484222325", 16);
    private static final BigInteger PRIME32 = new BigInteger("01000193", 16);
    private static final BigInteger PRIME64 = new BigInteger("100000001b3", 16);
    private static final BigInteger MOD32 = new BigInteger("2").pow(32);
    private static final BigInteger MOD64 = new BigInteger("2").pow(64);

    public static BigInteger fnv1_32(byte[] data) {
        BigInteger hash = INIT32;

        for (byte b : data) {
            hash = hash.multiply(PRIME32).mod(MOD32);
            hash = hash.xor(BigInteger.valueOf((int) b & 0xff));
        }

        return hash;
    }

    public static BigInteger fnv1_64(byte[] data) {
        BigInteger hash = INIT64;

        for (byte b : data) {
            hash = hash.multiply(PRIME64).mod(MOD64);
            hash = hash.xor(BigInteger.valueOf((int) b & 0xff));
        }

        return hash;
    }

    public static BigInteger fnv1a_32(byte[] data) {
        BigInteger hash = INIT32;

        for (byte b : data) {
            hash = hash.xor(BigInteger.valueOf((int) b & 0xff));
            hash = hash.multiply(PRIME32).mod(MOD32);
        }

        return hash;
    }

    public static BigInteger fnv1a_64(byte[] data) {
        BigInteger hash = INIT64;

        for (byte b : data) {
            hash = hash.xor(BigInteger.valueOf((int) b & 0xff));
            hash = hash.multiply(PRIME64).mod(MOD64);
        }

        return hash;
    }
}