#pragma once
#include <sstream>
#include <vector>
#include <cinttypes>
#include <rapidjson/rapidjson.h>
#include <rapidjson/document.h>
#include <rapidjson/prettywriter.h>
#include <rapidjson/writer.h>

#define JSON_GET(json, storage, ...) storage = json.get<decltype(storage)>(__VA_ARGS__)

namespace util
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
    JSON(JSON& other);
    ~JSON() = default;

    /* Parses the document. Returns true if no parse errors */
    bool parse(const std::string& data);

    /* Sets the member with the specified value */
    template <typename T, typename... Args>
    void set(T&& value, Args&&... args);

    /* Retrieves the memebr with the specified value */
    template <typename T, typename... Args>
    T get(Args&&... args);

    /* Retrieves the element at the specified index, be sure to check if the document is an array before calling */
    template <typename T>
    T at(size_t index);

    /* Checks if the member exists */
    template <typename... Args>
    bool memberExists(Args&&... args);

    /* Erases the specifed member in the path*/
    template <typename... Args>
    bool erase(std::string mem, Args&&... args);

    /* Only if the document is an array, push the values to it, returns true if the document is an array, false otherwise */
    template <typename T, typename... Args>
    bool push(T&& value, Args&&... values);

    /* Sets the document to an empty array */
    void setArray();

    /* Sets the document to an empty object */
    void setObject();

    /* Returns true if the document is an array, false otherwise */
    bool isArray();

    /* Returns true if the document is an object, false otherwise */
    bool isObject();

    /* Iterates over each element if the document is an array and returns true, simply returns false if not an array */
    template <typename Callback>
    bool foreach (Callback function);

    /* Outputs the document as a compressed string */
    std::string toString();

    /* Outputs the document as a formatted string */
    std::string toPrettyString();

    std::string err();

    JSON& operator=(JSON& other);

   private:
    rapidjson::Document mDoc;
    std::string mErr;

    template <typename T>
    void setValue(rapidjson::Value* member, T&& value);

    template <size_t Size>
    void setValue(rapidjson::Value* member, char const (&value)[Size]);

    template <typename T>
    T getValue(rapidjson::Value* member);

    template <size_t Size>
    rapidjson::Value* getOrCreateMember(const char* path[Size]);

    template <size_t Size>
    rapidjson::Value* getMember(const char* path[Size]);

    /* Recursivly push arguments back */
    template <typename T, typename... Args>
    void pushBack(T& value, Args&... args);

    /* Pushes back a single element */
    template <typename T>
    void pushBack(T& value);

    /* Pushes a single string literal */
    template <size_t Size>
    void pushBack(char const (&value)[Size]);
  };

  inline JSON::JSON()
  {
    setObject();  // default to object type
  }

  inline JSON::JSON(JSON& other)
  {
    *this = other;
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

  template <typename T, typename... Args>
  void JSON::set(T&& value, Args&&... args)
  {
    const char* path[sizeof...(args)] = {args...};
    auto member = getOrCreateMember<sizeof...(args)>(path);
    assert(member != nullptr);
    setValue(member, value);
  }

  template <typename T, typename... Args>
  T JSON::get(Args&&... args)
  {
    const char* path[sizeof...(args)] = {args...};
    auto member = getMember<sizeof...(args)>(path);
    assert(member != nullptr);
    return getValue<T>(member);
  }

  template <typename... Args>
  bool JSON::memberExists(Args&&... args)
  {
    const char* path[sizeof...(args)] = {args...};
    return getMember<sizeof...(args)>(path) != nullptr;
  }

  template <typename... Args>
  bool JSON::erase(std::string mem, Args&&... args)
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

  template <typename T, typename... Args>
  inline bool JSON::push(T&& value, Args&&... values)
  {
    if (mDoc.IsArray()) {
      pushBack(value, values...);
      return true;
    }

    return false;
  }

  inline void JSON::setArray()
  {
    mDoc.SetArray();
  }

  inline void JSON::setObject()
  {
    mDoc.SetObject();
  }

  inline bool JSON::isArray()
  {
    return mDoc.IsArray();
  }

  inline bool JSON::isObject()
  {
    return mDoc.IsObject();
  }

  template <typename Callback>
  inline bool JSON::foreach (Callback function)
  {
    if (mDoc.IsArray()) {
      for (auto i = mDoc.Begin(); i != mDoc.End(); i++) {
        function(*i);
      }

      return true;
    }

    return false;
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

  inline JSON& JSON::operator=(JSON& other)
  {
    if (other.isArray()) {
      mDoc.SetArray();
    } else {
      mDoc.SetObject();
    }

    mDoc.Swap(other.mDoc);

    return *this;
  }

  /* Setters */

  template <>
  inline void JSON::setValue(rapidjson::Value* member, std::string& str)
  {
    member->SetString(rapidjson::StringRef(str.c_str()), mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const char*& str)
  {
    member->SetString(rapidjson::StringRef(str), mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, int& i)
  {
    member->SetInt(i);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, bool& b)
  {
    member->SetBool(b);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, float& f)
  {
    member->SetFloat(f);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, rapidjson::Value*& value)
  {
    member->SetObject();
    *member = rapidjson::Value(*value, mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, Object& object)
  {
    (void)object;
    member->SetObject();
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, Array& array)
  {
    (void)array;
    member->SetArray();
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint8_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint16_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint32_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint64_t& value)
  {
    member->SetUint64(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, JSON& other)
  {
    if (other.isArray()) {
      member->SetArray();
    } else {
      member->SetObject();
    }

    member->Swap(other.mDoc);
  }

  template <size_t Size>
  void JSON::setValue(rapidjson::Value* member, char const (&value)[Size])
  {
    member->SetString(rapidjson::StringRef(value, Size), mDoc.GetAllocator());
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
  inline bool JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsBool() ? member->GetBool() : false;
  }

  template <>
  inline float JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsFloat() ? member->GetFloat() : 0.0f;
  }

  template <>
  inline uint8_t JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsUint() ? member->GetUint() : 0;
  }

  template <>
  inline uint16_t JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsUint() ? member->GetUint() : 0;
  }

  template <>
  inline uint32_t JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsUint() ? member->GetUint() : 0;
  }

  template <>
  inline uint64_t JSON::getValue(rapidjson::Value* member)
  {
    return member && member->IsUint64() ? member->GetUint() : 0;
  }

  template <>
  inline JSON JSON::getValue(rapidjson::Value* member)
  {
    JSON doc;

    if (member != nullptr) {
      member->Swap(doc.mDoc);
    }

    return doc;
  }

  template <typename T>
  T JSON::at(size_t index)
  {
    return mDoc[index].Get<T>();
  }

  template <size_t Size>
  rapidjson::Value* JSON::getOrCreateMember(const char* path[Size])
  {
    rapidjson::Value* val = &mDoc;
    for (size_t i = 0; i < Size; i++) {
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

  template <size_t Size>
  rapidjson::Value* JSON::getMember(const char* path[Size])
  {
    rapidjson::Value* val = &mDoc;
    for (size_t i = 0; i < Size; i++) {
      auto& str = path[i];
      if (val->GetType() != rapidjson::Type::kObjectType || !val->HasMember(str)) {
        return nullptr;
      }
      val = &(*val)[str];
    }
    return val;
  }

  template <typename T, typename... Args>
  inline void JSON::pushBack(T& value, Args&... values)
  {
    pushBack(value);

    if constexpr (sizeof...(values) > 0) {
      pushBack(values...);
    }
  }

  template <typename T>
  inline void JSON::pushBack(T& value)
  {
    rapidjson::Value v;
    v.Set(value);
    mDoc.PushBack(v, mDoc.GetAllocator());
  }

  template <>
  inline void JSON::pushBack(JSON& value)
  {
    mDoc.PushBack(value.mDoc, mDoc.GetAllocator());
  }

  template <>
  inline void JSON::pushBack(const char*& value)
  {
    rapidjson::Value v;
    v.SetString(rapidjson::StringRef(value), mDoc.GetAllocator());
    mDoc.PushBack(v, mDoc.GetAllocator());
  }

  template <size_t Size>
  inline void JSON::pushBack(char const (&value)[Size])
  {
    rapidjson::Value v;
    v.SetString(rapidjson::StringRef(value, Size), mDoc.GetAllocator());
    mDoc.PushBack(v, mDoc.GetAllocator());
  }
}  // namespace json
