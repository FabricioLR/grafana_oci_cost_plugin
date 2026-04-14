# OCI Cost Visualizer - Grafana Plugin

Este é um plugin de backend para o Grafana desenvolvido em **Go**, projetado para permitir a visualização detalhada e o monitoramento de custos da **Oracle Cloud Infrastructure (OCI)** diretamente nos seus dashboards.

## 🚀 Por que este projeto?

A gestão financeira (FinOps) é um pilar crítico da infraestrutura em nuvem. Este projeto nasceu da necessidade de centralizar a observabilidade, permitindo que administradores e desenvolvedores acompanhem o consumo financeiro com a mesma facilidade que acompanham métricas de CPU ou Memória, utilizando filtros avançados e dados em tempo real.

## 🛠️ Tecnologias Utilizadas

- **Go (Golang):** Performance e eficiência no processamento de dados.
- **OCI SDK for Go:** Integração oficial com as APIs de Usage e Metering da Oracle.
- **Grafana Plugin SDK:** Para uma experiência nativa dentro do ecossistema Grafana.

## ✨ Funcionalidades

- **Filtros Avançados:** Filtre seus custos por:
  - Nome do Serviço (Service Name)
  - Namespace
  - Tags (Key/Value)
- **Granularidade Temporal:** Detalhamento por dia ou agrupamento mensal.
- **Otimização de Performance:** Implementação de **cache interno** para reduzir o tempo de resposta e evitar o rate limit da API da OCI.
- **Segurança:** Gestão segura de credenciais através do backend, suportando autenticação via API Keys e configuração de Tenancy OCID.

## 🔒 Segurança

O plugin foi desenvolvido seguindo as premissas de segurança do Grafana. As credenciais da OCI (como `private_key` e `fingerprint`) são armazenadas e processadas exclusivamente no lado do servidor (**Secure JSON Data**), garantindo que dados sensíveis nunca sejam transmitidos para o navegador do usuário.

## 📋 Pré-requisitos

- Grafana >= 9.x
- Uma conta na Oracle Cloud (OCI) com permissões de leitura para `usage-reports` e `cost-management`.

## 🔧 Configuração

1. Clone este repositório na pasta de plugins do seu Grafana.
2. Configure as variáveis de ambiente ou o arquivo de configuração da OCI (`~/.oci/config`).
3. No Grafana, adicione o Data Source "OCI Cost Visualizer".
4. Insira os detalhes da sua Tenancy, User OCID, Fingerprint e Private Key.