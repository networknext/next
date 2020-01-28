#ifndef ENCODING_BASE_STREAM_HPP
#define ENCODING_BASE_STREAM_HPP
namespace encoding
{
    /**
        Reads bit packed integer values from a buffer.
        Relies on the user reconstructing the exact same set of bit reads as bit writes when the buffer was written. This is an
       unattributed bitpacked binary stream! Implementation: 32 bit dwords are read in from memory to the high bits of a scratch
       value as required. The user reads off bit values from the scratch value from the right, after which the scratch value is
       shifted by the same number of bits.
     */

    /**
        Functionality common to all stream classes.
     */

    class BaseStream
    {
       public:
        /**
            Base stream constructor.
         */
        explicit BaseStream();

        /**
            Set a context on the stream.
         */

        void SetContext(void* context);

        /**
            Get the context pointer set on the stream.
            @returns The context pointer. May be NULL.
         */

        void* GetContext() const;

       private:
        void* m_context;  ///< The context pointer set on the stream. May be NULL.
    };
}  // namespace encoding
#endif