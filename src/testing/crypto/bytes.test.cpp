#include "includes.h"
#include "testing/test.hpp"
#include "crypto/bytes.hpp"

TEST(crypto_random_bytes)
{
  std::array<uint8_t, 32> buff1, buff2;
  CHECK(crypto::RandomBytes(buff1, sizeof(buff1)));
  CHECK(crypto::RandomBytes(buff2, sizeof(buff2)));
  CHECK(buff1 != buff2);
}

TEST(crypto_create_nonce_bytes)
{
  std::array<uint8_t, 32> buff1, buff2;
  CHECK(crypto::CreateNonceBytes(buff1));
  CHECK(crypto::CreateNonceBytes(buff2));
  CHECK(buff1 != buff2);
}

TEST(crypto_box)
{
  const std::array<uint8_t, 4> CRYPTO_BOX_MESSAGE = {'t', 'e', 's', 't'};
  const auto CRYPTO_BOX_CIPHERTEXT_LEN = CRYPTO_BOX_MESSAGE.size() + crypto_box_MACBYTES;

  std::array<unsigned char, crypto_box_PUBLICKEYBYTES> sender_publickey;
  std::array<unsigned char, crypto_box_SECRETKEYBYTES> sender_secretkey;
  CHECK(crypto_box_keypair(sender_publickey.data(), sender_secretkey.data()) == 0);

  std::array<unsigned char, crypto_box_PUBLICKEYBYTES> receiver_publickey;
  std::array<unsigned char, crypto_box_SECRETKEYBYTES> receiver_secretkey;
  crypto_box_keypair(receiver_publickey.data(), receiver_secretkey.data());

  std::array<unsigned char, crypto_box_NONCEBYTES> nonce;
  std::array<unsigned char, CRYPTO_BOX_CIPHERTEXT_LEN> ciphertext;
  CHECK(crypto::RandomBytes(nonce, nonce.size()));

  CHECK(
   crypto_box_easy(
    ciphertext.data(),
    CRYPTO_BOX_MESSAGE.data(),
    CRYPTO_BOX_MESSAGE.size(),
    nonce.data(),
    receiver_publickey.data(),
    sender_secretkey.data()) == 0);

  std::array<uint8_t, 4> decrypted;

  CHECK(
   crypto_box_open_easy(
    decrypted.data(), ciphertext.data(), ciphertext.size(), nonce.data(), sender_publickey.data(), receiver_secretkey.data()) ==
   0);

  CHECK(decrypted == CRYPTO_BOX_MESSAGE);
}

TEST(crypto_secret_box)
{
  const std::array<uint8_t, 4> CRYPTO_SECRET_BOX_MESSAGE = {'t', 'e', 's', 't'};
  const auto CRYPTO_SECRET_BOX_CIPHERTEXT_LEN = CRYPTO_SECRET_BOX_MESSAGE.size() + crypto_secretbox_MACBYTES;

  std::array<unsigned char, crypto_secretbox_KEYBYTES> key;
  std::array<unsigned char, crypto_secretbox_NONCEBYTES> nonce;
  std::array<unsigned char, CRYPTO_SECRET_BOX_CIPHERTEXT_LEN> ciphertext;

  crypto_secretbox_keygen(key.data());
  CHECK(crypto::CreateNonceBytes(nonce));
  CHECK(
   crypto_secretbox_easy(
    ciphertext.data(), CRYPTO_SECRET_BOX_MESSAGE.data(), CRYPTO_SECRET_BOX_MESSAGE.size(), nonce.data(), key.data()) == 0);

  std::array<unsigned char, 4> decrypted;
  CHECK(crypto_secretbox_open_easy(decrypted.data(), ciphertext.data(), ciphertext.size(), nonce.data(), key.data()) == 0);
  CHECK(decrypted == CRYPTO_SECRET_BOX_MESSAGE);
}

TEST(crypto_aead)
{
  const std::array<unsigned char, 4> CRYPTO_AEAD_MESSAGE = {'t', 'e', 's', 't'};
  const std::array<unsigned char, 6> CRYPTO_AEAD_ADDITIONAL_DATA = {'1', '2', '3', '4', '5', '6'};

  std::array<unsigned char, crypto_aead_chacha20poly1305_NPUBBYTES> nonce;
  std::array<unsigned char, crypto_aead_chacha20poly1305_KEYBYTES> key;
  std::array<unsigned char, CRYPTO_AEAD_MESSAGE.size() + crypto_aead_chacha20poly1305_ABYTES> ciphertext;
  unsigned long long ciphertext_len;

  crypto_aead_chacha20poly1305_keygen(key.data());
  CHECK(crypto::RandomBytes(nonce, nonce.size()));

  crypto_aead_chacha20poly1305_encrypt(
   ciphertext.data(),
   &ciphertext_len,
   CRYPTO_AEAD_MESSAGE.data(),
   CRYPTO_AEAD_MESSAGE.size(),
   CRYPTO_AEAD_ADDITIONAL_DATA.data(),
   CRYPTO_AEAD_ADDITIONAL_DATA.size(),
   nullptr,
   nonce.data(),
   key.data());

  std::array<unsigned char, CRYPTO_AEAD_MESSAGE.size()> decrypted;
  unsigned long long decrypted_len;
  CHECK(
   crypto_aead_chacha20poly1305_decrypt(
    decrypted.data(),
    &decrypted_len,
    nullptr,
    ciphertext.data(),
    ciphertext.size(),
    CRYPTO_AEAD_ADDITIONAL_DATA.data(),
    CRYPTO_AEAD_ADDITIONAL_DATA.size(),
    nonce.data(),
    key.data()) == 0);
}

TEST(crypto_aead_ietf)
{
  const std::array<uint8_t, 4> CRYPTO_AEAD_IETF_MESSAGE = {'t', 'e', 's', 't'};
  const std::array<uint8_t, 6> CRYPTO_AEAD_IETF_ADDITIONAL_DATA = {'1', '2', '3', '4', '5', '6'};

  std::array<unsigned char, crypto_aead_xchacha20poly1305_ietf_NPUBBYTES> nonce;
  std::array<unsigned char, crypto_aead_xchacha20poly1305_ietf_KEYBYTES> key;
  std::array<unsigned char, CRYPTO_AEAD_IETF_MESSAGE.size() + crypto_aead_xchacha20poly1305_ietf_ABYTES> ciphertext;
  unsigned long long ciphertext_len;

  crypto_aead_xchacha20poly1305_ietf_keygen(key.data());
  CHECK(crypto::CreateNonceBytes(nonce));

  crypto_aead_xchacha20poly1305_ietf_encrypt(
   ciphertext.data(),
   &ciphertext_len,
   CRYPTO_AEAD_IETF_MESSAGE.data(),
   CRYPTO_AEAD_IETF_MESSAGE.size(),
   CRYPTO_AEAD_IETF_ADDITIONAL_DATA.data(),
   CRYPTO_AEAD_IETF_ADDITIONAL_DATA.size(),
   nullptr,
   nonce.data(),
   key.data());

  std::array<unsigned char, CRYPTO_AEAD_IETF_MESSAGE.size()> decrypted;
  unsigned long long decrypted_len;
  CHECK(
   crypto_aead_xchacha20poly1305_ietf_decrypt(
    decrypted.data(),
    &decrypted_len,
    nullptr,
    ciphertext.data(),
    ciphertext.size(),
    CRYPTO_AEAD_IETF_ADDITIONAL_DATA.data(),
    CRYPTO_AEAD_IETF_ADDITIONAL_DATA.size(),
    nonce.data(),
    key.data()) == 0);
}

TEST(crypto_sign)
{
  const std::array<uint8_t, 4> CRYPTO_SIGN_MESSAGE = {'t', 'e', 's', 't'};

  std::array<unsigned char, crypto_sign_PUBLICKEYBYTES> public_key;
  std::array<unsigned char, crypto_sign_SECRETKEYBYTES> private_key;
  crypto_sign_keypair(public_key.data(), private_key.data());

  std::array<unsigned char, CRYPTO_SIGN_MESSAGE.size() + crypto_sign_BYTES> signed_message;
  unsigned long long signed_message_len;

  CHECK(
   crypto_sign(
    signed_message.data(), &signed_message_len, CRYPTO_SIGN_MESSAGE.data(), CRYPTO_SIGN_MESSAGE.size(), private_key.data()) ==
   0);

  std::array<unsigned char, CRYPTO_SIGN_MESSAGE.size()> unsigned_message;
  unsigned long long unsigned_message_len;

  CHECK(
   crypto_sign_open(
    unsigned_message.data(), &unsigned_message_len, signed_message.data(), signed_message_len, public_key.data()) == 0);
}

TEST(crypto_sign_detached)
{
  const std::array<uint8_t, 22> MESSAGE_PART1 = {
   'A', 'r', 'b', 'i', 't', 'r', 'a', 'r', 'y', ' ', 'd', 'a', 't', 'a', ' ', 't', 'o', ' ', 'h', 'a', 's', 'h',
  };
  const std::array<uint8_t, 23> MESSAGE_PART2 = {
   'i', 's', ' ', 'l', 'o', 'n', 'g', 'e', 'r', ' ', 't', 'h', 'a', 'n', ' ', 'e', 'x', 'p', 'e', 'c', 't', 'e', 'd',
  };

  std::array<unsigned char, crypto_sign_PUBLICKEYBYTES> public_key;
  std::array<unsigned char, crypto_sign_SECRETKEYBYTES> private_key;
  CHECK(crypto_sign_keypair(public_key.data(), private_key.data()) == 0);

  std::array<unsigned char, crypto_sign_BYTES> signature;

  crypto_sign_state state;

  CHECK(crypto_sign_init(&state) == 0);
  CHECK(crypto_sign_update(&state, MESSAGE_PART1.data(), MESSAGE_PART1.size()) == 0);
  CHECK(crypto_sign_update(&state, MESSAGE_PART2.data(), MESSAGE_PART2.size()) == 0);
  CHECK(crypto_sign_final_create(&state, signature.data(), nullptr, private_key.data()) == 0);

  CHECK(crypto_sign_init(&state) == 0);
  CHECK(crypto_sign_update(&state, MESSAGE_PART1.data(), MESSAGE_PART1.size()) == 0);
  CHECK(crypto_sign_update(&state, MESSAGE_PART2.data(), MESSAGE_PART2.size()) == 0);
  CHECK(crypto_sign_final_verify(&state, signature.data(), public_key.data()) == 0);
}

TEST(crypto_key_exchange)
{
  std::array<uint8_t, crypto_kx_PUBLICKEYBYTES> client_public_key;
  std::array<uint8_t, crypto_kx_SECRETKEYBYTES> client_private_key;
  CHECK(crypto_kx_keypair(client_public_key.data(), client_private_key.data()) == 0);

  std::array<uint8_t, crypto_kx_PUBLICKEYBYTES> server_public_key;
  std::array<uint8_t, crypto_kx_SECRETKEYBYTES> server_private_key;
  CHECK(crypto_kx_keypair(server_public_key.data(), server_private_key.data()) == 0);

  std::array<uint8_t, crypto_kx_SESSIONKEYBYTES> client_send_key;
  std::array<uint8_t, crypto_kx_SESSIONKEYBYTES> client_receive_key;
  CHECK(
   crypto_kx_client_session_keys(
    client_receive_key.data(),
    client_send_key.data(),
    client_public_key.data(),
    client_private_key.data(),
    server_public_key.data()) == 0);

  std::array<uint8_t, crypto_kx_SESSIONKEYBYTES> server_send_key;
  std::array<uint8_t, crypto_kx_SESSIONKEYBYTES> server_receive_key;
  CHECK(
   crypto_kx_server_session_keys(
    server_receive_key.data(),
    server_send_key.data(),
    server_public_key.data(),
    server_private_key.data(),
    client_public_key.data()) == 0);

  CHECK(client_send_key == server_receive_key);
  CHECK(server_send_key == client_receive_key);
}
