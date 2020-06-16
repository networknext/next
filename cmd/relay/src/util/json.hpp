#pragma once

namespace util
{
  class JSON
  {
   public:
    enum class Type : uint8_t
    {
      Null = rapidjson::Type::kNullType,
      Object = rapidjson::Type::kObjectType,
      Array = rapidjson::Type::kArrayType,
      Number = rapidjson::Type::kNumberType,
      String = rapidjson::Type::kStringType,
      False = rapidjson::Type::kFalseType,
      True = rapidjson::Type::kTrueType
    };

    JSON();
    JSON(JSON& other);
    ~JSON() = default;

    /* Parses the document. Returns true if no errors.
     * Can be anything that supplies a char* through T::data() and size through T::size();
     */
    template <typename T>
    bool parse(const T& data);

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
    template <typename Callback>  // template so lambdas are used directly and inlined rather than wrapped in a std::function
    bool foreach(Callback function);

    /* Returns true if the specified member's underlying type is correct, false otherwise, even if the member does not exist */
    template <typename... Args>
    bool memberIs(Type t, Args&&... args);

    /* Outputs the document as a compressed string */
    std::string toString();

    /* Outputs the document as a formatted string */
    std::string toPrettyString();

    /* Returns an internal errors if there is one */
    std::string err();

    /* Returns the internal document */
    rapidjson::Document& internal();

    JSON& operator=(JSON& other);

    rapidjson::Value& operator[](size_t index);
    rapidjson::Value& operator[](std::string member);

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

  inline bool operator==(JSON::Type t1, rapidjson::Type t2)
  {
    return static_cast<uint8_t>(t1) == static_cast<uint8_t>(t2);
  }

  inline bool operator==(rapidjson::Type t1, JSON::Type t2)
  {
    return t2 == t1;
  }

  inline JSON::JSON()
  {
    setObject();  // default to object type
  }

  inline JSON::JSON(JSON& other)
  {
    *this = other;
  }

  template <typename T>
  inline bool JSON::parse(const T& raw)
  {
    rapidjson::ParseResult result = mDoc.Parse(raw.data(), raw.size());

    if (!result && mDoc.HasParseError()) {
      std::stringstream errorStream;
      errorStream << "Document parse error: " << mDoc.GetParseError();
      mErr = errorStream.str();
      return false;
    }

    return true;
  }

  template <typename T, typename... Args>
  void JSON::set(T&& value, Args&&... args)
  {
    const char* path[sizeof...(args)] = {args...};
    auto member = getOrCreateMember<sizeof...(args)>(path);
    setValue(member, value);
  }

  template <typename T, typename... Args>
  T JSON::get(Args&&... args)
  {
    const char* path[sizeof...(args)] = {args...};
    auto member = getMember<sizeof...(args)>(path);
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
  inline bool JSON::foreach(Callback function)
  {
    if (mDoc.IsArray()) {
      for (auto i = mDoc.Begin(); i != mDoc.End(); i++) {
        function(*i);
      }

      return true;
    }

    return false;
  }

  template <typename... Args>
  bool JSON::memberIs(JSON::Type t, Args&&... args)
  {
    const char* path[sizeof...(args)] = {args...};
    const auto size = sizeof...(args);
    rapidjson::Value* val = &mDoc;

    for (size_t i = 0; i < size; i++) {
      auto& str = path[i];

      if (!val->HasMember(str)) {
        return false;
      }

      val = &(*val)[str];
    }

    return val->GetType() == t;
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

  inline rapidjson::Document& JSON::internal()
  {
    return mDoc;
  }

  inline JSON& JSON::operator=(JSON& other)
  {
    mDoc = std::move(other.mDoc);
    return *this;
  }

  inline rapidjson::Value& JSON::operator[](size_t index)
  {
    return mDoc[index];
  }

  inline rapidjson::Value& JSON::operator[](std::string member)
  {
    return mDoc[member.c_str()];
  }

  /* Setters */

  template <>
  inline void JSON::setValue(rapidjson::Value* member, JSON& other)
  {
    member->CopyFrom(other.mDoc, mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, std::string& str)
  {
    member->SetString(rapidjson::StringRef(str.c_str()), mDoc.GetAllocator());
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const std::string& str)
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
  inline void JSON::setValue(rapidjson::Value* member, const int& i)
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
  inline void JSON::setValue(rapidjson::Value* member, uint8_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const uint8_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint16_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const uint16_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint32_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const uint32_t& value)
  {
    member->SetUint(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, uint64_t& value)
  {
    member->SetUint64(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const uint64_t& value)
  {
    member->SetUint64(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, double& value)
  {
    member->SetDouble(value);
  }

  template <>
  inline void JSON::setValue(rapidjson::Value* member, const double& value)
  {
    member->SetDouble(value);
  }

  template <size_t Size>
  inline void JSON::setValue(rapidjson::Value* member, char const (&value)[Size])
  {
    member->SetString(rapidjson::StringRef(value, Size), mDoc.GetAllocator());
  }

  /* Getters */

  template <>
  inline JSON JSON::getValue(rapidjson::Value* member)
  {
    JSON doc;

    if (member != nullptr) {
      doc.mDoc.CopyFrom(*member, doc.mDoc.GetAllocator());
    } else {
      doc.mDoc.SetNull();
    }

    return doc;
  }

  template <>
  inline rapidjson::Value* JSON::getValue(rapidjson::Value* member)
  {
    return member;
  }

  template <>
  inline std::string JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kStringType) ? std::string(member->GetString()) : std::string();
  }

  template <>
  inline const char* JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kStringType) ? member->GetString() : "";
  }

  template <>
  inline int JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<int>() : 0;
  }

  template <>
  inline int64_t JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->IsInt64()) ? member->GetInt64() : 0;
  }

  template <>
  inline bool JSON::getValue(rapidjson::Value* member)
  {
    return (member && (member->GetType() == rapidjson::Type::kTrueType || member->GetType() == rapidjson::Type::kFalseType))
            ? member->Get<bool>()
            : false;
  }

  template <>
  inline float JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<float>() : 0.0f;
  }

  template <>
  inline double JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<double>() : 0.0;
  }

  template <>
  inline uint8_t JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<uint32_t>() : 0;
  }

  template <>
  inline uint16_t JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<uint32_t>() : 0;
  }

  template <>
  inline uint32_t JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<uint32_t>() : 0;
  }

  template <>
  inline uint64_t JSON::getValue(rapidjson::Value* member)
  {
    return (member && member->GetType() == rapidjson::Type::kNumberType) ? member->Get<uint64_t>() : 0;
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

  template <>
  inline void JSON::pushBack(JSON& value)
  {
    rapidjson::Value tmp;
    tmp.CopyFrom(value.mDoc, mDoc.GetAllocator());
    mDoc.PushBack(tmp, mDoc.GetAllocator());
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

  template <typename T>
  inline void JSON::pushBack(T& value)
  {
    rapidjson::Value v;
    v.Set(value);
    mDoc.PushBack(v, mDoc.GetAllocator());
  }
}  // namespace ftypes
