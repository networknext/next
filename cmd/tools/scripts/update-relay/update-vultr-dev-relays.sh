#!/bin/bash

export SSH_KEY

ips=(
    144.202.31.156
    45.76.58.249
    45.63.51.158
    149.28.53.151
    149.248.32.18
    149.28.213.194
)

public_keys=(
    6S16CNisuSKxbwyUntcFSeJycX+JgYvmxZ+A8RL2s1U=
    6HX1kUXiBoSl+iZa19YfGaldsSyGOoiSFaQEmfYqbgo=
    We6Q/v37+nkYw5JO+ZTP/X58BHMvJFBMIC/MdbfhpCA=
    mVsdDDEfVOYgQl+yeyzapnto5K35pnh+qSSS6HepNU4=
    0fUcWSHRIWOBShgwnO2Iw0rTuTV0bvTMIusj6o0eSnk=
    PAVXbAT2NGUvkZxjtebBwrEI5x+siQ0RfpN6ddoEHV4=
)

private_keys=(
    hI2OVFQoxqXPR6CHd23PwlLi7kwSfCipmskVt2nEgqg=
    Obi/iXitNVwtfnxIsigEGrjhxjjWV0eu/wQ3+eA9uqE=
    gzh1fWLv8BbFNGy6wbXYJ9u00dCXHHDeCzaCR1QQhLQ=
    F/hv3gTJpea/hFaIyRFn48ZGYFQjTJUHWklbDVdSzzg=
    9vjwHgHUOmZsqmw8hJB2k1toXDZt2NaBBdcCfYNoQUk=
    J5402JhmRAn/LNiI2ZDgKOY6HWBaQD7/6jascyv7v9Q=
)

for arg in "$@"
do
    case $arg in
        -f|--ssh-key)
        SSH_KEY="$2"
        shift
        shift
        ;;
    esac
done

if [ -z $SSH_KEY ]; then
    echo "No SSH key provided - use -f or --ssh_key to provide a file path to the SSH key"
    exit
fi

for i in ${!ips[@]}; do
    echo Updating relay at ip ${ips[$i]}
    ./cmd/tools/scripts/update-relay/update.sh -u root -i ${ips[$i]} -p ${public_keys[$i]} -s ${private_keys[$i]} -f $SSH_KEY
done
