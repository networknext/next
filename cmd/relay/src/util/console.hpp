#pragma once
#include <iostream>
#include <string>
#include <ctime>
#include <strings.h>

#define CODE_FORGROUND_RESET "\x1b[39m"
#define CODE_FORGROUND_BLACK "\x1b[30m"
#define CODE_FORGROUND_RED "\x1b[31m"
#define CODE_FORGROUND_GREEN "\x1b[32m"
#define CODE_FORGROUND_YELLOW "\x1b[33m"
#define CODE_FORGROUND_BLUE "\x1b[34m"
#define CODE_FORGROUND_MAGENTA "\x1b[35m"
#define CODE_FORGROUND_CYAN "\x1b[36m"
#define CODE_FORGROUND_WHITE "\x1b[37m"

#define CODE_BACKGROUND_RESET "\x1b[49m"
#define CODE_BACKGROUND_BLACK "\x1b[40m"
#define CODE_BACKGROUND_RED "\x1b[41m"
#define CODE_BACKGROUND_GREEN "\x1b[42m"
#define CODE_BACKGROUND_YELLOW "\x1b[43m"
#define CODE_BACKGROUND_BLUE "\x1b[44m"
#define CODE_BACKGROUND_MAGENTA "\x1b[45m"
#define CODE_BACKGROUND_CYAN "\x1b[46m"
#define CODE_BACKGROUND_WHITE "\x1b[47m"

#define CODE_FONTSTYLE_BOLD "\x1b[1m"
#define CODE_FONTSTYLE_FAINT "\x1b[2m"
#define CODE_FONTSTYLE_ITALIC "\x1b[3m"
#define CODE_FONTSTYLE_UNDERLINE "\x1b[4m"
#define CODE_FONTSTYLE_UNDERLINE_OFF "\x1b[24m"
#define CODE_FONTSTYLE_CROSSEDOUT "\x1b[9m"
#define CODE_FONTSTYLE_CROSSEDOUT_OFF "\x1b[29m"

#define CODE_MISC_FULLRESET "\x1b[m"
#define CODE_MISC_SLOWBLINK "\x1b[5m"
#define CODE_MISC_BLINK_OFF "\x1b[25m"

namespace util
{
    enum class Mod
    {
        /* Forground */
        FG_Reset,
        FG_Black,
        FG_Red,
        FG_Green,
        FG_Yellow,
        FG_Blue,
        FG_Magenta,
        FG_Cyan,
        FG_White,

        /* Background */
        BG_Reset,
        BG_Black,
        BG_Red,
        BG_Green,
        BG_Yellow,
        BG_Blue,
        BG_Magenta,
        BG_Cyan,
        BG_White,

        /* Font Style */
        FS_Bold,
        FS_Faint,
        FS_Italic,
        FS_Underline,
        FS_UnderlineOff,
        FS_CrossedOut,
        FS_CrossedOutOff,

        /* Misc */
        M_FullReset,
        M_SlowBlink,
        M_BlinkOff
    };

    class Console
    {
       public:
        Console() = default;

        template <Mod E>
        constexpr const char* setOpt()
        {
            return tostr<E>().value;
        }

        template <typename... Args>
        void write(Args&&... args)
        {
            ((std::cout << std::forward<Args>(args)), ...) << setOpt<Mod::M_FullReset>();
        }

        template <typename... Args>
        void writeLine(Args&&... args)
        {
            write(args..., '\n');
        }

        template <typename... Args>
        void log(Args&&... args)
        {
            writeLine('[', StrTime(), "] ", args...);
        }

       private:
        static std::string StrTime();

        template <Mod E>
        struct tostr
        {
            const char* const value = nullptr;
        };
    };

    inline std::string Console::StrTime()
    {
        char timebuff[16];
        bzero(timebuff, sizeof(timebuff));
        auto t = time(nullptr);
        auto timestruct = localtime(&t);
        std::strftime(timebuff, sizeof(timebuff) - 1, "%I:%M:%S %P", timestruct);
        return std::string(timebuff);
    }

    /* Forground */
    template <>
    struct Console::tostr<Mod::FG_Reset>
    {
        const char* const value = CODE_FORGROUND_RESET;
    };

    template <>
    struct Console::tostr<Mod::FG_Black>
    {
        const char* const value = CODE_FORGROUND_BLACK;
    };

    template <>
    struct Console::tostr<Mod::FG_Red>
    {
        const char* const value = CODE_FORGROUND_RED;
    };

    template <>
    struct Console::tostr<Mod::FG_Green>
    {
        const char* const value = CODE_FORGROUND_GREEN;
    };

    template <>
    struct Console::tostr<Mod::FG_Yellow>
    {
        const char* const value = CODE_FORGROUND_YELLOW;
    };

    template <>
    struct Console::tostr<Mod::FG_Blue>
    {
        const char* const value = CODE_FORGROUND_BLUE;
    };

    template <>
    struct Console::tostr<Mod::FG_Magenta>
    {
        const char* const value = CODE_FORGROUND_MAGENTA;
    };

    template <>
    struct Console::tostr<Mod::FG_Cyan>
    {
        const char* const value = CODE_FORGROUND_CYAN;
    };

    template <>
    struct Console::tostr<Mod::FG_White>
    {
        const char* const value = CODE_FORGROUND_WHITE;
    };

    /* Background */
    template <>
    struct Console::tostr<Mod::BG_Reset>
    {
        const char* const value = CODE_BACKGROUND_RESET;
    };

    template <>
    struct Console::tostr<Mod::BG_Black>
    {
        const char* const value = CODE_BACKGROUND_BLACK;
    };

    template <>
    struct Console::tostr<Mod::BG_Red>
    {
        const char* const value = CODE_BACKGROUND_RED;
    };

    template <>
    struct Console::tostr<Mod::BG_Green>
    {
        const char* const value = CODE_BACKGROUND_GREEN;
    };

    template <>
    struct Console::tostr<Mod::BG_Yellow>
    {
        const char* const value = CODE_BACKGROUND_YELLOW;
    };

    template <>
    struct Console::tostr<Mod::BG_Blue>
    {
        const char* const value = CODE_BACKGROUND_BLUE;
    };

    template <>
    struct Console::tostr<Mod::BG_Magenta>
    {
        const char* const value = CODE_BACKGROUND_MAGENTA;
    };

    template <>
    struct Console::tostr<Mod::BG_Cyan>
    {
        const char* const value = CODE_BACKGROUND_CYAN;
    };

    template <>
    struct Console::tostr<Mod::BG_White>
    {
        const char* const value = CODE_BACKGROUND_WHITE;
    };

    /* Font styles */
    template <>
    struct Console::tostr<Mod::FS_Bold>
    {
        const char* const value = CODE_FONTSTYLE_BOLD;
    };

    template <>
    struct Console::tostr<Mod::FS_Faint>
    {
        const char* const value = CODE_FONTSTYLE_FAINT;
    };

    template <>
    struct Console::tostr<Mod::FS_Italic>
    {
        const char* const value = CODE_FONTSTYLE_ITALIC;
    };

    template <>
    struct Console::tostr<Mod::FS_Underline>
    {
        const char* const value = CODE_FONTSTYLE_UNDERLINE;
    };

    template <>
    struct Console::tostr<Mod::FS_UnderlineOff>
    {
        const char* const value = CODE_FONTSTYLE_UNDERLINE_OFF;
    };

    template <>
    struct Console::tostr<Mod::FS_CrossedOut>
    {
        const char* const value = CODE_FONTSTYLE_CROSSEDOUT;
    };

    template <>
    struct Console::tostr<Mod::FS_CrossedOutOff>
    {
        const char* const value = CODE_FONTSTYLE_CROSSEDOUT_OFF;
    };

    /* Misc */
    template <>
    struct Console::tostr<Mod::M_FullReset>
    {
        const char* const value = CODE_MISC_FULLRESET;
    };

    template <>
    struct Console::tostr<Mod::M_SlowBlink>
    {
        const char* const value = CODE_MISC_SLOWBLINK;
    };

    template <>
    struct Console::tostr<Mod::M_BlinkOff>
    {
        const char* const value = CODE_MISC_BLINK_OFF;
    };
}  // namespace dash
