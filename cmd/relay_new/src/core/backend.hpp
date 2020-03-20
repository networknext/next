#ifndef CORE_BACKEND_HPP
#define CORE_BACKEND_HPP
namespace core
{
  class Backend
  {
   public:
    bool init();
    bool update();

   private:

   CURL* mCurl;
  };
}  // namespace core
#endif