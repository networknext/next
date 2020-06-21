package com.networknext.dataflow.util;

import java.nio.charset.StandardCharsets;

import com.networknext.api.id.Id.EntityId;
import com.networknext.dataflow.util.fnv.FNV;

public class EntityIdHelpers {

    public static String getStringForNetworkNextStorage(EntityId id) {
        if (id == null) {
            return null;
        }
        return id.getKind() + "/" + id.getName();
    }

    public static String getStringForThirdPartyStorage(EntityId id) {
        if (id == null) {
            return "0";
        }
        return FNV.fnv1a_64(id.getName().getBytes(StandardCharsets.UTF_8)).toString(10);
    }

}