#!/bin/bash

# Array of language codes
languages=(fr it ja pt-BR pt-PT)

# Array of translation mappings (key:translation)
declare -A translations

# Settings translations
translations["legalAndPrivacy"]="Legal & Privacy|Légal et Confidentialité|Legale e Privacy|法的事項とプライバシー|Legal e Privacidade|Legal e Privacidade"
translations["privacyPolicy"]="Privacy Policy|Politique de Confidentialité|Informativa sulla Privacy|プライバシーポリシー|Política de Privacidade|Política de Privacidade"
translations["termsOfService"]="Terms of Service|Conditions d'Utilisation|Termini di Servizio|利用規約|Termos de Serviço|Termos de Serviço"

# Signup translations
translations["agreeToTerms"]="I agree to the Terms of Service and Privacy Policy|J'accepte les Conditions d'Utilisation et la Politique de Confidentialité|Accetto i Termini di Servizio e l'Informativa sulla Privacy|利用規約とプライバシーポリシーに同意します|Concordo com os Termos de Serviço e Política de Privacidade|Concordo com os Termos de Serviço e Política de Privacidade"
translations["mustAgreeToTerms"]="You must agree to the Terms of Service and Privacy Policy to continue|Vous devez accepter les Conditions d'Utilisation et la Politique de Confidentialité pour continuer|Devi accettare i Termini di Servizio e l'Informativa sulla Privacy per continuare|続行するには利用規約とプライバシーポリシーに同意する必要があります|Você deve concordar com os Termos de Serviço e Política de Privacidade para continuar|Deve concordar com os Termos de Serviço e Política de Privacidade para continuar"

# Data export translations  
translations["dataExportTitle"]="Export My Data|Exporter Mes Données|Esporta i Miei Dati|データをエクスポート|Exportar Meus Dados|Exportar os Meus Dados"
translations["dataExportMessage"]="You can request a copy of all your data. We'll send it to your registered email address within 30 days.|Vous pouvez demander une copie de toutes vos données. Nous l'enverrons à votre adresse e-mail enregistrée dans les 30 jours.|Puoi richiedere una copia di tutti i tuoi dati. Te la invieremo al tuo indirizzo email registrato entro 30 giorni.|すべてのデータのコピーをリクエストできます。30日以内に登録されたメールアドレスに送信します。|Você pode solicitar uma cópia de todos os seus dados. Enviaremos para seu endereço de email registrado em até 30 dias.|Pode solicitar uma cópia de todos os seus dados. Enviaremos para o seu endereço de email registado no prazo de 30 dias."
translations["dataExportRequestButton"]="Request Data Export|Demander l'Export de Données|Richiedi Esportazione Dati|データエクスポートをリクエスト|Solicitar Exportação de Dados|Solicitar Exportação de Dados"
translations["dataExportEmailSubject"]="Data Export Request|Demande d'Export de Données|Richiesta di Esportazione Dati|データエクスポートリクエスト|Solicitação de Exportação de Dados|Solicitação de Exportação de Dados"
translations["dataExportEmailBody"]="I would like to request a copy of all my data associated with the email: {{email}}|Je souhaiterais demander une copie de toutes mes données associées à l'e-mail : {{email}}|Vorrei richiedere una copia di tutti i miei dati associati all'email: {{email}}|メール {{email}} に関連付けられたすべてのデータのコピーをリクエストしたいと思います|Gostaria de solicitar uma cópia de todos os meus dados associados ao email: {{email}}|Gostaria de solicitar uma cópia de todos os meus dados associados ao email: {{email}}"

echo "This script would update translations for all languages"
echo "For now, please manually update the remaining language files with the English translations as fallbacks"