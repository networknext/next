<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <h3 class="card-header">Monitoring Service Evaluation Terms</h3>
          <div class="card-body" style="overflow-y: scroll;">
            NETWORK NEXT<br />
            MONITORING SERVICE EVALUATION TERMS<br />
            <br />
            <p>
              This document summarizes the terms of the self-deployed monitoring service (the Service) provided by Network Next. By accepting the terms and conditions or using the service, you (the Service Consumer) agree to be bound by the terms. The Service may be updated or terminated by Network Next at its sole discretion at any time for any reason.
            </p>
            <p>
              Network Next hereby grants to Service Consumer, a limited, non-exclusive, non-transferable and non-sublicensable license during the Term to access and utilize the Service for evaluation purposes, subject to these terms. Each Party has the limited right to use the other Party's trademarks to publicize their collaboration and use of the Service.
            </p>
            <p>
              Network Next anonymizes any personal data during collection and before it is transmitted and stored by Network Next, by removing all direct and indirect identifiers, and on which technical, organizational, and legal controls reasonably prevent re-identifying individuals so that it is no longer Personal Data. The collected data which is anonymized consists of the user's IP address and their user id (e.g, Steam ID, PSN ID) (the "Anonymized Data").
            </p>
            <p>
              Network Next also collects other aggregated, non-personally identifiable internet usage analytics, including the route calculated for the user to reach the server, when and how long the user is connected to the server, RTT, jitter, and packet loss statistics about the user's connection, broad geographical data inferred based on IP Address, and the user's ISP name and their ISP's ASN (the "Technical Data"). No Personal Data is contained in the Technical Data.
            </p>
            <p>
              The Service Consumer hereby grants to Network Next the irrevocable, worldwide, perpetual, royalty-free right to access, use, reproduce, distribute, display, assign, sell, license, sublicense, and create derivative works from Technical Data and Anonymized Data, for any purpose (to the extent permitted by law) at Network Next's sole discretion, provided that any such usage is not directly attributable to the Service Consumer or its individual users.
            </p>
            <p>
              Each Party represents that it (i) has the requisite legal authority and ability to enter into this Agreement and perform according to its terms, (ii) has no legal or other restrictions preventing it from entering into, performing, or paying compensation under this Agreement, as applicable, and (ii) will comply with all laws and regulations applicable to its provision, receipt, or use of the Services, as applicable.
              EXCEPT FOR THE FOREGOING, NETWORK NEXT MAKES NO REPRESENTATIONS OR WARRANTIES AND THE SERVICE CONSUMER HAS NOT RELIED ON ANY REPRESENTATION OR WARRANTY, EXPRESS OR IMPLIED, REGARDING THE SERVICES. SERVICE CONSUMER AGREES THAT THEY ARE LICENSING THE SERVICES ON AN "AS IS" BASIS, WITHOUT ANY EXPRESS, IMPLIED, OR STATUTORY WARRANTIES. NETWORK NEXT MAKES NO PROMISES AS TO THE INCLUSION OF ANY PARTICULAR FEATURES OR FUNCTIONALITY OF THE SERVICES, WHICH MAY BE MODIFIED BY NETWORK NEXT THROUGHOUT THE TERM, EXCEPT THAT NETWORK NEXT WARRANTS, AT A MINIMUM, THAT THE CURRENTLY-AVAILABLE FUNCTIONALITY OF THE SERVICES AS OF THE DATE OF THIS AGREEMENT (OR A REASONABLE SUBSTITUTE) WILL BE AVAILABLE THROUGHOUT THE TERM.<br />
              Each Party (the "Indemnifying Party") will defend the other Party (the "Indemnified Party"), and indemnify them against any damages, claims, penalties, costs, and expenses (including attorneys' fees) to the extent arising from an actual or alleged breach by the Indemnifying Party of their representations under this Agreement, provided that the Indemnified Party promptly alerts the Indemnifying Party of the event giving rise to indemnity and provides reasonable cooperation with the Indemnifying Party in the defense or settlement of any claim.
            </p>
            <p>
              Neither Party will have any liability or be responsible to the other Party, or any third party, for any indirect, incidental, consequential, special, punitive or other damages, including lost opportunity, revenue, profits, or the cost of procurement of substitute services in any way arising out of or related to the use of the Service, even if a Party has been advised of the possibility of such damages.
              The relationship between the Parties is that of independent contractors. This Agreement is not authority for one Party to act as agent or to make commitments for the other Party.
              Neither Party may assign this Agreement without the other Party's prior written consent, except that no consent is required if either Party sells all or substantially all of its assets.
              The prevailing party in any proceeding arising out of or related to these terms shall receive as additional damages reasonable attorney's fees and costs.
              If any provision of these term is found illegal, unenforceable or invalid, the remaining provisions shall nevertheless be carried into full force and effect.
              These terms represent the entire agreement of the Parties with respect to the Service and cannot be modified in any way unless in writing signed by the Parties.
              These terms are governed by the laws of the state of New York. Any proceeding arising out of or related to this Agreement will be brought exclusively in the federal or state courts located in New York, NY, and the Parties hereby consent to the jurisdiction of said courts in any proceeding arising out of or related to this Agreement.
            </p>
          </div>
          <div class="card-footer">
            <div class="row">
              <div class="col" style="text-align: center;">
                <button class="btn btn-xs btn-secondary" style="max-height: 40px; width: 100px;" @click="$root.$emit('hideTOSModal', false)">
                  Deny
                </button>
              </div>
              <div class="col-2">
              </div>
              <div class="col" style="text-align: center;">
                <button class="btn btn-xs btn-success" style="max-height: 40px; width: 100px;" @click="$root.$emit('hideTOSModal', true)">
                  Accept
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'

/**
 * This component opens up a modal for picking from nested map points
 */

/**
 * TODO: Clean up template
 */

@Component
export default class TermsOfServiceModal extends Vue {}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .modal-mask {
    position: fixed;
    z-index: 9998;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: table;
    transition: opacity 0.3s ease;
  }

  .modal-wrapper {
    display: table-cell;
    vertical-align: middle;
  }

  .modal-container {
    max-width: 800px;
    max-height: 600px;
    margin: 0px auto 8%;
    background-color: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
    transition: all 0.3s ease;
  }

  .modal-header h3 {
    margin-top: 0;
    color: #42b983;
  }

  .modal-body {
    margin: 20px 0;
  }

  .modal-default-button {
    float: right;
    border-color: #009FDF;
    background-color: #009FDF;
  }

  .modal-default-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }

  /*
  * The following styles are auto-applied to elements with
  * transition="modal" when their visibility is toggled
  * by Vue.js.
  *
  * You can easily play with the modal transition by editing
  * these styles.
  */

  .modal-enter {
    opacity: 0;
  }

  .modal-leave-active {
    opacity: 0;
  }

  .modal-enter .modal-container,
  .modal-leave-active .modal-container {
    -webkit-transform: scale(1.1);
    transform: scale(1.1);
  }

  .my-custom-scrollbar {
    position: relative;
    height: 300px;
    overflow: auto;
  }
  .table-wrapper-scroll-y {
    display: block;
  }
</style>
