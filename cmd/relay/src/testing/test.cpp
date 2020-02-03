#include "test.hpp"

#include "config.hpp"
#include "macros.hpp"
#include "encoding/read.test.hpp"
#include "encoding/write.test.hpp"
#include "net/address.test.hpp"
#include "legacy.hpp"

#include "relay/relay.hpp"

namespace testing
{
    void relay_test()
    {
        printf("\nRunning relay tests:\n\n");

        check(relay::relay_initialize() == RELAY_OK);

        RUN_TEST(TestRead);
        RUN_TEST(TestWrite);
        RUN_TEST(TestAddress);
        RUN_TEST(TestLegacy);
        printf("\n");

        fflush(stdout);

        relay::relay_term();
    }
}  // namespace testing