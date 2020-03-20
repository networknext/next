#pragma once
#include <sstream>
#include <vector>
#include <rapidjson/rapidjson.h>
#include <rapidjson/document.h>
#include <rapidjson/prettywriter.h>
#include <rapidjson/writer.h>

#define JSON_GET(json, storage, ...) storage = json.get<decltype(storage)>(__VA_ARGS__)

namespace json
{
  /* Entire purpose is to be able to set a field as an array while keeping the rapidjson abstraction */
  class Array
  {};

  /* Entire purpose is to be able to set a field as an object while keeping the rapidjson abstraction */
  class Object
  {};

  class JSON
  {
   public:
    JSON();
    ~JSON() = default;

    /* Parses the document. Returns true if no parse errors */
    bool parse(const std::string& data);

    /* Sets the member with the specified value */
    template <typename T, typename... Args>
    void set(T value, Args&&... args)
    {
      const char* path[sizeof...(args)] = {args...};
      auto member = getOrCreateMember<sizeof...(args)>(path);
      assert(member != nullptr);
      setValue(member, value);
    }

    /* Retrieves the memebr with the specified value */
    template <typename T, typename... Args>
    T get(Args&&... args)
    {
      const char* path[sizeof...(args)] = {args...};
      auto member = getMember<sizeof...(args)>(path);
      assert(member != nullptr);
      return getValue<T>(member);
    }

    /* Checks if the member exists */
    template <typename... Args>
    bool memberExists(Args&&... args)
    {
      const char* path[sizeof...(args)] = {args...};
      return getMember<sizeof...(args)>(path) != nullptr;
    }

    /* Erases the specifed member in the path*/
    template <typename... Args>
    bool erase(std::string mem, Args&&... args)
    {
      const char* path[sizeof...(args)] = {args...};
      auto member = getMember<sizeof...(args)>(path);
      assert(member != nullptr);
      if (member->HasMember(mem.c_str())) {
        member->EraseMember(mem.c_str());
        return true;
      } else {
        return false;
      }
    }

    /* Outputs the document as a compressed string */
    std::string toString();

    /* Outputs the document as a formatted string */
    std::string toPrettyString();

    std::string err();

   private:
    rapidjson::Document mDoc;
    std::string mErr;

    template <typename T>
    void setValue(rapidjson::Value* member, T value);

    template <typename T>
    T getValue(rapidjson::Value* member);

    template <size_t size>
    rapidjson::Value* getOrCreateMember(const char* path[size])
    {
      rapidjson::Value* val = &mDoc;
      for (size_t i = 0; i < size; i++) {
        auto& str = path[i];

        if (val->GetType() != rapidjson::Type::kObjectType) {
          val->SetObject();
        }

        if (!val->HasMember(str)) {
          val->AddMember(rapidjson::StringRef(str), rapidjson::Value(rapidjson::kNullType), mDoc.GetAllocator());
        }

        val = &(*val)[str];
      }
      return val;
    }

    template <size_t size>
    rapidjson::Value* getMember(const char* path[size])
    {
      rapidjson::Value* val = &mDoc;
      for (size_t i = 0; i < size; i++) {
        auto& str = path[i];
        if (val->GetType() != rapidjson::Type::kObjectType || !val->HasMember(str)) {
          return nullptr;
        }
        val = &(*val)[str];
      }
      return val;
    }
  };

  inline JSON::JSON()
  {
    mDoc.SetObject();  // default document to type object
  }

  inline bool JSON::parse(const std::string& data)
  {
    rapidjson::ParseResult result = mDoc.Parse(data.c_str());
    std::stringstream error_stream;

    if (!result && mDoc.HasParseError()) {
      error_stream << "Document parse error: " << mDoc.GetParseError();
      mErr = error_stream.str();
      return false;
    }

    return true;
  }

  inline std::string JSON::toString()
  {
    rapidjson::StringBuffer buff;
    rapidjson::Writer<rapidjson::StringBuffer> writer(buff);
    mDoc.Accept(writer);

    return buff.GetString();
  }

  inline std::string JSON::toPrettyString()
  {
    rapidjson::StringBuffer buff;
    rapidjson::PrettyWriter<rapidjson::StringBuffer> pwriter(buff);
    mDoc.Accept(pwriter);

    return buff.GetString();
  }

  inline std::string JSON::err()
  {
    return mErr;
  }

  /* Setters */

  template <>
  inline void JSON::setValue(rapidjson::Value* member, std::string str)
  {
    member->SetString(rapidjson::StringRef(str.c_str()), mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const char* str)
  {
    member->SetString(rapidjson::StringRef(str), mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, int i)
  {
    member->SetInt(i);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, unsigned int i)
  {
    member->SetUint(i);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, bool b)
  {
    member->SetBool(b);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, float f)
  {
    member->SetFloat(f);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, rapidjson::Value* value)
  {
    member->SetObject();
    *member = rapidjson::Value(*value, mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, Object object)
  {
    (void)object;
    member->SetObject();
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, Array array)
  {
    (void)array;
    member->SetArray();
  }

  /* Getters */

  template <>
  inline rapidjson::Value* JSON::getValue(rapidjson::Value* member)
  {
    return member;
  }

  template <>
  inline std::string JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsString() ? std::string(member->GetString()) : std::string();
  }

  template <>
  inline const char* JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsString() ? member->GetString() : "";
  }

  template <>
  inline int JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsInt() ? member->GetInt() : 0;
  }

  template <>
  inline unsigned int JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsUint() ? member->GetUint() : 0;
  }

  template <>
  inline bool JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsBool() ? member->GetBool() : false;
  }

  template <>
  inline float JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsFloat() ? member->GetFloat() : 0.0f;
  }
}  // namespace json
